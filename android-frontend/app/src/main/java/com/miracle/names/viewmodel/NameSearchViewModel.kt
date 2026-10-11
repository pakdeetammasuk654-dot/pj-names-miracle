package com.miracle.names.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.miracle.names.data.model.NameRerankRequest
import com.miracle.names.data.model.NameResultItem
import com.miracle.names.data.model.NameSearchRequest
import com.miracle.names.data.remote.RetrofitClient
import kotlinx.coroutines.CancellationException
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.SharingStarted
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.combine
import kotlinx.coroutines.flow.stateIn
import kotlinx.coroutines.launch

sealed class UiState {
    object Idle : UiState()
    object Loading : UiState()
    data class Success(
        val results: List<NameResultItem>,
        val snapshot: AppliedSearchSnapshot
    ) : UiState()

    data class Error(val message: String) : UiState()
}

private const val SUPPORTED_RANKING_CONTRACT = "adaptive_pairs_v1"

class NameSearchViewModel : ViewModel() {

    private val _uiState = MutableStateFlow<UiState>(UiState.Idle)
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    // The results screen is dismissed back to the form after a successful search.
    // Keep this completion state independent of uiState so Step 4 remains complete.
    private val _hasCompletedSearch = MutableStateFlow(false)
    val hasCompletedSearch: StateFlow<Boolean> = _hasCompletedSearch.asStateFlow()

    private val _connectionError = MutableStateFlow<String?>(null)
    val connectionError: StateFlow<String?> = _connectionError.asStateFlow()

    private val _isCheckingConnection = MutableStateFlow(true)
    val isCheckingConnection: StateFlow<Boolean> = _isCheckingConnection.asStateFlow()

    val searchMode = MutableStateFlow(MODE_STANDALONE)
    val isFindSurnameMode = MutableStateFlow(false)
    val surname = MutableStateFlow("")
    val dayOfBirth = MutableStateFlow("")
    val prototypeName = MutableStateFlow("")
    private val _nameSuggestions = MutableStateFlow<List<NameResultItem>>(emptyList())
    val nameSuggestions: StateFlow<List<NameResultItem>> = _nameSuggestions.asStateFlow()
    private val _suggestionPreviewPairs = MutableStateFlow<List<com.miracle.names.data.model.SuggestionPairPreview>>(emptyList())
    val suggestionPreviewPairs: StateFlow<List<com.miracle.names.data.model.SuggestionPairPreview>> = _suggestionPreviewPairs.asStateFlow()
    private val _suggestionCount = MutableStateFlow(0)
    val suggestionCount: StateFlow<Int> = _suggestionCount.asStateFlow()
    private val _suggestionsUnlocked = MutableStateFlow(false)
    val suggestionsUnlocked: StateFlow<Boolean> = _suggestionsUnlocked.asStateFlow()
    private val _isLoadingNameSuggestions = MutableStateFlow(false)
    val isLoadingNameSuggestions: StateFlow<Boolean> = _isLoadingNameSuggestions.asStateFlow()
    private val _selectedSuggestionName = MutableStateFlow("")
    val selectedSuggestionName: StateFlow<String> = _selectedSuggestionName.asStateFlow()
    val selectedWish = MutableStateFlow(DEFAULT_WISH)

    private val _pendingRankingWeights = MutableStateFlow(RankingWeights.Default)
    val pendingRankingWeights: StateFlow<RankingWeights> = _pendingRankingWeights.asStateFlow()

    private var searchJob: Job? = null
    private val searchRevisionGate = RequestRevisionGate()
    private var candidateToken = ""
    private val _isReranking = MutableStateFlow(false)
    val isReranking: StateFlow<Boolean> = _isReranking.asStateFlow()
    private val _rerankError = MutableStateFlow<String?>(null)
    val rerankError: StateFlow<String?> = _rerankError.asStateFlow()
    val rankMovements = MutableStateFlow<Map<Int, Int>>(emptyMap())
    val newlyRankedIds = MutableStateFlow<Set<Int>>(emptySet())
    val rerankRevision = MutableStateFlow(0)
    private var rerankJob: Job? = null
    private var movementClearJob: Job? = null
    private var rerankRequestRevision = 0L

    // Selected item for Blessing Card View.
    val selectedItem = MutableStateFlow<NameResultItem?>(null)
    val showGuideSheet = MutableStateFlow(false)
    val articles = MutableStateFlow<List<com.miracle.names.data.model.ArticleItem>>(emptyList())
    val isLoadingArticles = MutableStateFlow(false)

    // Lifetime In-App Purchase unlocked state (Google Play Billing + Local Backup)
    var billingManager: com.miracle.names.data.billing.BillingManager? = null

    private val _realIsUnlocked = MutableStateFlow(false)
    private val _mockPurchaseEnabled = MutableStateFlow(false)
    val mockPurchaseEnabled: StateFlow<Boolean> = _mockPurchaseEnabled.asStateFlow()
    val isUnlocked: StateFlow<Boolean> = combine(_realIsUnlocked, _mockPurchaseEnabled) { real, mock ->
        real || mock
    }.stateIn(viewModelScope, SharingStarted.Eagerly, false)

    fun initializeBilling(context: android.content.Context) {
        if (billingManager == null) {
            val manager = com.miracle.names.data.billing.BillingManager(context.applicationContext)
            billingManager = manager
            viewModelScope.launch {
                manager.isUnlocked.collect { unlocked ->
                    val wasUnlocked = _realIsUnlocked.value
                    _realIsUnlocked.value = unlocked
                    if (unlocked && !wasUnlocked) {
                        if (prototypeName.value.isNotBlank() && _selectedSuggestionName.value.isBlank()) {
                            requestNameSuggestions(prototypeName.value)
                        }
                        val currentUi = _uiState.value
                        if (currentUi is UiState.Success) {
                            searchNames()
                        }
                    }
                }
            }
            viewModelScope.launch {
                manager.mockPurchaseEnabled.collect { enabled ->
                    _mockPurchaseEnabled.value = enabled
                }
            }
        }
    }

    fun setMockPurchaseEnabled(enabled: Boolean) {
        _mockPurchaseEnabled.value = enabled
        billingManager?.setMockPurchaseEnabled(enabled)
    }

    fun unlockAllNames(activity: android.app.Activity? = null) {
        if (activity != null && billingManager != null) {
            val manager = billingManager
            if (manager?.hasProductDetails() == true) {
                manager.launchPurchaseFlow(activity)
            } else {
                manager?.startConnection()
            }
        }
    }

    fun restorePurchases() {
        billingManager?.restorePurchases()
    }

    fun getDisplayName(fullName: String, isFree: Boolean = false): String {
        if (isUnlocked.value || isFree) return fullName
        if (fullName.isBlank()) return fullName
        val visibleLength = (fullName.length / 3).coerceAtLeast(2)
        val visiblePart = fullName.take(visibleLength)
        return "$visiblePart***"
    }

    // Celebrities list.
    val celebrities = MutableStateFlow<List<com.miracle.names.data.model.Celebrity>>(emptyList())

    // Decoded pairs from backend.
    val decodedPairs = MutableStateFlow<List<String>>(emptyList())
    val decodedPairTypes = MutableStateFlow<List<String>>(emptyList())
    val decodedPairPoints = MutableStateFlow<List<Int>>(emptyList())
    val decodedPairIsGood = MutableStateFlow<List<Boolean>>(emptyList())
    val decodedPairDescs = MutableStateFlow<List<String>>(emptyList())
    val decodedKalakini = MutableStateFlow(com.miracle.names.data.model.KalakiniInfo())
    val decodedProfile4D = MutableStateFlow<com.miracle.names.data.model.NameProfile4D?>(null)
    val decodedMeaning = MutableStateFlow("")
    val isDecodedFromDb = MutableStateFlow(false)
    private val _pinnedWishMeaning = MutableStateFlow("")
    val pinnedWishMeaning: StateFlow<String> = _pinnedWishMeaning.asStateFlow()
    private val _hasUserSelectedWishMeaning = MutableStateFlow(false)
    val hasUserSelectedWishMeaning: StateFlow<Boolean> = _hasUserSelectedWishMeaning.asStateFlow()
    private val _meaningPool = MutableStateFlow<List<String>>(emptyList())
    val meaningPool: StateFlow<List<String>> = _meaningPool.asStateFlow()
    private val _meaningPageIndex = MutableStateFlow(0)
    val meaningPageIndex: StateFlow<Int> = _meaningPageIndex.asStateFlow()
    private val _isMeaningAutoPlaying = MutableStateFlow(true)
    val isMeaningAutoPlaying: StateFlow<Boolean> = _isMeaningAutoPlaying.asStateFlow()
    private val _randomMeaningSuggestions = MutableStateFlow<List<String>>(DEFAULT_AUSPICIOUS_MEANINGS.shuffled().take(3))
    val randomMeaningSuggestions: StateFlow<List<String>> = _randomMeaningSuggestions.asStateFlow()
    private val _isAutoSelectingMeaning = MutableStateFlow(false)
    val isAutoSelectingMeaning: StateFlow<Boolean> = _isAutoSelectingMeaning.asStateFlow()
    private val _isMeaningAutoSelected = MutableStateFlow(false)
    val isMeaningAutoSelected: StateFlow<Boolean> = _isMeaningAutoSelected.asStateFlow()
    private val _isDecodingName = MutableStateFlow(false)
    val isDecodingName: StateFlow<Boolean> = _isDecodingName.asStateFlow()
    private val _decodeError = MutableStateFlow<String?>(null)
    val decodeError: StateFlow<String?> = _decodeError.asStateFlow()
    private var decodeJob: Job? = null
    private var suggestionJob: Job? = null
    private var decodeRequestRevision = 0L

    fun updatePrototypeNameInput(value: String) {
        prototypeName.value = value
        _selectedSuggestionName.value = ""
        decodeCurrentName(value)
        requestNameSuggestions(value)
    }

    fun selectSuggestedName(name: String) {
        val cleanName = name.trim()
        prototypeName.value = cleanName
        _selectedSuggestionName.value = cleanName
        decodeCurrentName(cleanName)
        requestNameSuggestions(cleanName)
    }

    private fun requestNameSuggestions(query: String) {
        suggestionJob?.cancel()
        _nameSuggestions.value = emptyList()
        _suggestionPreviewPairs.value = emptyList()
        _suggestionCount.value = 0
        _suggestionsUnlocked.value = false
        if (query.trim().length < 2) {
            _nameSuggestions.value = emptyList()
            _isLoadingNameSuggestions.value = false
            return
        }
        _isLoadingNameSuggestions.value = true
        suggestionJob = viewModelScope.launch {
            try {
                delay(280)
                val response = RetrofitClient.apiService.suggestNames(query.trim())
                val body = response.body()
                if (response.isSuccessful && body != null && !body.error && prototypeName.value.trim() == query.trim()) {
                    _nameSuggestions.value = body.data
                    _suggestionPreviewPairs.value = body.previewPairs
                    _suggestionCount.value = body.count
                    _suggestionsUnlocked.value = body.premiumUnlocked
                }
            } catch (_: CancellationException) {
                throw CancellationException()
            } catch (_: Exception) {
                _nameSuggestions.value = emptyList()
            } finally {
                if (prototypeName.value.trim() == query.trim()) _isLoadingNameSuggestions.value = false
            }
        }
    }

    fun totalMeaningPages(): Int {
        val pool = _meaningPool.value
        return if (pool.isEmpty()) 36 else pool.size / 3
    }

    fun loadRandomMeanings(pinFirst: Boolean) {
        viewModelScope.launch {
            try {
                val meanings = fetchRandomMeanings()
                if (meanings.isNotEmpty()) {
                    // Ensure exactly up to 108 meanings (36 sets of 3)
                    val exactCount = minOf(meanings.size, 108) - (minOf(meanings.size, 108) % 3)
                    _meaningPool.value = meanings.take(exactCount)
                    _meaningPageIndex.value = 0
                    updatePageSuggestions()
                } else if (_meaningPool.value.isEmpty()) {
                    _meaningPool.value = DEFAULT_AUSPICIOUS_MEANINGS.shuffled()
                    _meaningPageIndex.value = 0
                    updatePageSuggestions()
                }
            } catch (_: CancellationException) {
                throw CancellationException()
            } catch (_: Exception) {
                if (_meaningPool.value.isEmpty()) {
                    _meaningPool.value = DEFAULT_AUSPICIOUS_MEANINGS.shuffled()
                    _meaningPageIndex.value = 0
                    updatePageSuggestions()
                }
            }
        }
    }

    private fun updatePageSuggestions() {
        val pool = _meaningPool.value
        if (pool.isEmpty()) {
            _randomMeaningSuggestions.value = DEFAULT_AUSPICIOUS_MEANINGS.shuffled().take(3)
            return
        }
        val total = totalMeaningPages()
        val currentIndex = _meaningPageIndex.value.coerceIn(0, (total - 1).coerceAtLeast(0))
        val start = currentIndex * 3
        val end = minOf(start + 3, pool.size)
        if (start < end) {
            _randomMeaningSuggestions.value = pool.subList(start, end)
        }
    }

    fun nextMeaningPage() {
        val total = totalMeaningPages()
        if (total > 0) {
            _meaningPageIndex.value = (_meaningPageIndex.value + 1) % total
            updatePageSuggestions()
        }
    }

    fun previousMeaningPage() {
        val total = totalMeaningPages()
        if (total > 0) {
            _meaningPageIndex.value = (_meaningPageIndex.value - 1 + total) % total
            updatePageSuggestions()
        }
    }

    fun toggleMeaningAutoPlay() {
        _isMeaningAutoPlaying.value = !_isMeaningAutoPlaying.value
    }

    private suspend fun fetchRandomMeanings(): List<String> {
        val response = RetrofitClient.apiService.getRandomMeanings(limit = 108)
        if (!response.isSuccessful) return emptyList()
        return response.body()?.meanings
            .orEmpty()
            .map(String::trim)
            .filter(String::isNotEmpty)
            .distinct()
    }

    fun pinWishMeaning(meaning: String) {
        val cleanMeaning = meaning
            .removePrefix("✨")
            .removePrefix("💡")
            .trim()
        if (cleanMeaning.isBlank()) return
        _pinnedWishMeaning.value = cleanMeaning
        selectedWish.value = ""
        _hasUserSelectedWishMeaning.value = true
        _isMeaningAutoSelected.value = !isDecodedFromDb.value
        _isMeaningAutoPlaying.value = false
    }

    val selectedPairDetail = MutableStateFlow<com.miracle.names.data.model.PairDetailResponse?>(null)
    val isPairDetailLoading = MutableStateFlow(false)

    fun fetchPairDetail(pairNumber: String) {
        val cleanPair = pairNumber.trim()
        if (cleanPair.isBlank()) return

        isPairDetailLoading.value = true
        viewModelScope.launch {
            try {
                val response = RetrofitClient.apiService.getPairDetail(cleanPair)
                if (response.isSuccessful && response.body()?.error == false) {
                    selectedPairDetail.value = response.body()
                } else {
                    selectedPairDetail.value = com.miracle.names.data.model.PairDetailResponse(
                        pairNumber = cleanPair,
                        pairType = "D10",
                        miracleDesc = "คู่เลขมงคลสิริมงคลประจำนาม",
                        miracleDetail = "คู่เลข $cleanPair ถือเป็นคู่เลขที่มีพลังมงคลโดดเด่น ส่งเสริมให้เกิดความสำเร็จ ความมั่นคง และสิ่งดีงามในชีวิตตามตำราเลขศาสตร์",
                        isGood = true
                    )
                }
            } catch (_: Exception) {
                selectedPairDetail.value = com.miracle.names.data.model.PairDetailResponse(
                    pairNumber = cleanPair,
                    pairType = "D10",
                    miracleDesc = "คู่เลขมงคลสิริมงคลประจำนาม",
                    miracleDetail = "คู่เลข $cleanPair ถือเป็นคู่เลขที่มีพลังมงคลโดดเด่น ส่งเสริมให้เกิดความสำเร็จ ความมั่นคง และสิ่งดีงามในชีวิตตามตำราเลขศาสตร์",
                    isGood = true
                )
            } finally {
                isPairDetailLoading.value = false
            }
        }
    }

    fun dismissPairDetail() {
        selectedPairDetail.value = null
        isPairDetailLoading.value = false
    }

    companion object {
        const val DEFAULT_WISH = "🌸 เสริมดวงมงคลรอบด้าน (ทุกมิติชีวิต)"
        const val CONNECTION_ERROR_MESSAGE = "ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้ กรุณาตรวจสอบการเชื่อมต่อแล้วลองใหม่"
        val DEFAULT_AUSPICIOUS_MEANINGS = listOf(
            "ความสำเร็จ ความเจริญก้าวหน้าในหน้าที่การงาน",
            "มีเสน่ห์ เมตตามหานิยม เป็นที่รักใคร่ของผู้คน",
            "ร่ำรวย มั่งคั่ง มั่นคง มีโชคลาภวาสนาบารมี",
            "สติปัญญาเฉลียวฉลาด รอบรู้ มีไหวพริบปฏิภาณ",
            "สุขภาพแข็งแรง อายุยืนยาว แคล้วคลาดปลอดภัย",
            "ชีวิตราบรื่น มีความสุข สมหวังในสิ่งที่ปรารถนา"
        )
    }

    init {
        fetchCelebrities()
        fetchArticles()
        loadRandomMeanings(pinFirst = false)
    }

    fun updateSearchMode(mode: String) {
        val resolvedMode = if (mode == MODE_COMBINED) MODE_COMBINED else MODE_STANDALONE
        if (resolvedMode == searchMode.value) return

        searchMode.value = resolvedMode
        isFindSurnameMode.value = false
        // Clear input when switching mode so user can type afresh (e.g. surname)
        prototypeName.value = ""
        surname.value = ""
        decodeCurrentName("")
    }

    fun toggleCombinedSearchTarget() {
        if (searchMode.value != MODE_COMBINED) return
        isFindSurnameMode.value = !isFindSurnameMode.value
        // The fixed input changes role between surname and first name. Reusing
        // the old text would silently reinterpret a first name as a surname (or
        // vice versa), so require an explicit value for the newly selected role.
        surname.value = ""
        decodeCurrentName("")
    }

    fun updateRankingWeights(numerology: Float, semantic: Float, sound: Float, similarity: Float) {
        _pendingRankingWeights.value = normalizeRankingWeights(numerology, semantic, sound, similarity)
    }

    fun updateRankingWeightsAndRerank(numerology: Float, semantic: Float, sound: Float, similarity: Float) {
        val requestedWeights = normalizeRankingWeights(numerology, semantic, sound, similarity)
        _pendingRankingWeights.value = requestedWeights

        val currentState = _uiState.value as? UiState.Success ?: return
        rerankJob?.cancel()
        val requestRevision = ++rerankRequestRevision

        if (candidateToken.isBlank()) {
            rollbackPendingWeights(
                currentState,
                "ไม่สามารถปรับอันดับได้ เนื่องจากชุดผลลัพธ์หมดอายุ กรุณาค้นหาใหม่"
            )
            return
        }

        _isReranking.value = true
        _rerankError.value = null
        val token = candidateToken
        rerankJob = viewModelScope.launch {
            try {
                delay(250)
                val response = RetrofitClient.apiService.rerankNames(
                    NameRerankRequest(
                        candidateToken = token,
                        numerologyWeight = requestedWeights.numerology,
                        semanticWeight = requestedWeights.semantic,
                        soundWeight = requestedWeights.sound,
                        similarityWeight = requestedWeights.similarity,
                        limit = 100
                    )
                )
                if (requestRevision != rerankRequestRevision) return@launch

                val body = response.body()
                if (response.isSuccessful && body != null && !body.error &&
                    body.policy.rankingContractVersion == SUPPORTED_RANKING_CONTRACT
                ) {
                    val latestState = _uiState.value as? UiState.Success ?: return@launch
                    val appliedMode = resolveAppliedMode(
                        body.appliedMode.ifBlank { body.mode },
                        latestState.snapshot.mode
                    )
                    val responseRequestedWeights = resolveResponseWeights(
                        body.requestedWeights,
                        requestedWeights
                    )
                    val responseAppliedWeights = resolveResponseWeights(
                        body.appliedWeights,
                        responseRequestedWeights
                    )
                    val safeData = body.data ?: emptyList()
                    val resultCount = resolveActualCount(
                        body.returnedCount.takeIf { it > 0 || safeData.isEmpty() } ?: body.count,
                        safeData.size
                    )
                    val nextSnapshot = latestState.snapshot.copy(
                        mode = appliedMode,
                        surname = resolveAppliedSurname(appliedMode, latestState.snapshot.surname),
                        weights = responseAppliedWeights,
                        requestedWeights = responseRequestedWeights,
                        count = resultCount,
                        eligibleCandidateCount = resolveEligibleCandidateCount(
                            body.eligibleCandidateCount,
                            resultCount
                        ),
                        semanticAvailable = body.semanticAvailable,
                        semanticSource = body.semanticSource,
                        semanticModel = body.resolvedSemanticQueryModel(),
                        semanticVectorHash = body.semanticVectorHash,
                        candidateVectorProvenanceVerified = body.candidateVectorProvenanceVerified,
                        candidateVectorProvenanceSource = body.candidateVectorProvenanceSource,
                        policy = body.policy,
                        appliedWeightsConfirmed = body.appliedWeights != null
                    )

                    val oldRanks = latestState.results.mapIndexed { index, item ->
                        item.nameId to resolvedRank(item, index)
                    }.toMap()
                    rankMovements.value = body.data.mapIndexedNotNull { index, item ->
                        val previousRank = oldRanks[item.nameId] ?: return@mapIndexedNotNull null
                        val movement = previousRank - resolvedRank(item, index)
                        if (movement == 0) null else item.nameId to movement
                    }.toMap()
                    newlyRankedIds.value = body.data.asSequence()
                        .map { it.nameId }
                        .filterNot(oldRanks::containsKey)
                        .toSet()

                    // Results and the exact weights/provenance that produced them change together.
                    _uiState.value = UiState.Success(body.data, nextSnapshot)
                    _pendingRankingWeights.value = responseRequestedWeights
                    selectedItem.value = selectedItem.value?.let { selected ->
                        body.data.firstOrNull { it.nameId == selected.nameId }
                    }
                    if (body.candidateToken.isNotBlank()) {
                        candidateToken = body.candidateToken
                    }
                    rerankRevision.value += 1
                    movementClearJob?.cancel()
                    movementClearJob = viewModelScope.launch {
                        delay(2200)
                        rankMovements.value = emptyMap()
                        newlyRankedIds.value = emptySet()
                    }
                } else {
                    val message = if (body != null && !body.error &&
                        body.policy.rankingContractVersion != SUPPORTED_RANKING_CONTRACT
                    ) {
                        "เวอร์ชันการจัดอันดับของระบบไม่ตรงกับแอป กรุณาอัปเดตแอปหรือค้นหาใหม่"
                    } else {
                        body?.message?.takeIf { it.isNotBlank() }
                            ?: "ไม่สามารถปรับอันดับได้ กรุณาลองใหม่"
                    }
                    rollbackPendingWeights(currentState, message)
                }
            } catch (cancelled: CancellationException) {
                throw cancelled
            } catch (_: Exception) {
                if (requestRevision == rerankRequestRevision) {
                    rollbackPendingWeights(
                        currentState,
                        "ไม่สามารถปรับอันดับได้ กรุณาตรวจสอบการเชื่อมต่อ"
                    )
                }
            } finally {
                if (requestRevision == rerankRequestRevision) {
                    _isReranking.value = false
                }
            }
        }
    }

    private fun rollbackPendingWeights(state: UiState.Success, message: String) {
        _pendingRankingWeights.value = state.snapshot.requestedWeights
        _rerankError.value = message
        _isReranking.value = false
    }

    fun decodeCurrentName(name: String) {
        decodeJob?.cancel()
        val requestRevision = ++decodeRequestRevision
        _decodeError.value = null
        clearDecodedName()

        if (name.isBlank()) {
            _isDecodingName.value = false
            selectedWish.value = DEFAULT_WISH
            return
        }

        val requestedName = name
        _isDecodingName.value = true
        decodeJob = viewModelScope.launch {
            try {
                // Avoid issuing one request for every keystroke while the user is typing.
                delay(220)
                val response = RetrofitClient.apiService.decodeName(requestedName, dayOfBirth.value)
                if (!isCurrentDecodeRequest(requestRevision, requestedName)) return@launch

                val body = response.body()
                if (response.isSuccessful && body != null && !body.error) {
                    decodedPairs.value = body.pairs.orEmpty()
                    decodedPairTypes.value = body.pairTypes.orEmpty()
                    decodedPairPoints.value = body.pairPoints.orEmpty()
                    decodedPairIsGood.value = body.pairIsGood.orEmpty()
                    decodedPairDescs.value = body.pairDescs.orEmpty()
                    decodedKalakini.value = body.kalakini
                    decodedProfile4D.value = body.profile4D
                    body.profile4D?.let { profile ->
                        val total = (profile.numerologyScore + profile.meaningScore +
                            profile.soundScore + profile.resonanceScore).coerceAtLeast(0.001f)
                        updateRankingWeights(
                            profile.numerologyScore / total,
                            profile.meaningScore / total,
                            profile.soundScore / total,
                            profile.resonanceScore / total
                        )
                    }
                    decodedMeaning.value = body.meaning.orEmpty().trim()
                    isDecodedFromDb.value = decodedMeaning.value.isNotBlank()
                    _pinnedWishMeaning.value = ""
                    _hasUserSelectedWishMeaning.value = false
                    _connectionError.value = null
                    _isMeaningAutoSelected.value = false
                    selectedWish.value = ""

                    val meanings = try {
                        fetchRandomMeanings()
                    } catch (cancelled: CancellationException) {
                        throw cancelled
                    } catch (_: Exception) {
                        emptyList()
                    }
                    if (isCurrentDecodeRequest(requestRevision, requestedName) && meanings.isNotEmpty()) {
                        _randomMeaningSuggestions.value = meanings.take(4)
                    }
                    if (isCurrentDecodeRequest(requestRevision, requestedName)) {
                        _isAutoSelectingMeaning.value = false
                    }
                } else {
                    onDecodeRejected(
                        requestRevision,
                        requestedName,
                        body?.message?.takeIf { it.isNotBlank() }
                            ?: "ชื่อนี้มีอักขระที่ระบบยังคำนวณไม่ได้"
                    )
                }
            } catch (cancelled: CancellationException) {
                throw cancelled
            } catch (_: Exception) {
                onDecodeFailure(requestRevision, requestedName)
            } finally {
                if (isCurrentDecodeRequest(requestRevision, requestedName)) {
                    _isDecodingName.value = false
                }
            }
        }
    }

    fun refreshSelectedNameForBirthday() {
        val name = prototypeName.value.trim()
        val day = dayOfBirth.value.trim()
        if (name.isBlank() || day.isBlank()) return
        viewModelScope.launch {
            try {
                val response = RetrofitClient.apiService.decodeName(name, day)
                val body = response.body()
                if (response.isSuccessful && body != null && !body.error) {
                    decodedKalakini.value = body.kalakini
                }
            } catch (_: Exception) { }
        }
    }

    private fun isCurrentDecodeRequest(requestRevision: Long, requestedName: String): Boolean {
        return requestRevision == decodeRequestRevision
    }

    private fun onDecodeFailure(requestRevision: Long, requestedName: String) {
        if (!isCurrentDecodeRequest(requestRevision, requestedName)) return
        clearDecodedName()
        _connectionError.value = CONNECTION_ERROR_MESSAGE
        selectedWish.value = DEFAULT_WISH
    }

    private fun onDecodeRejected(requestRevision: Long, requestedName: String, message: String) {
        if (!isCurrentDecodeRequest(requestRevision, requestedName)) return
        clearDecodedName()
        _decodeError.value = message
        selectedWish.value = DEFAULT_WISH
    }

    private fun clearDecodedName() {
        decodedPairs.value = emptyList()
        decodedPairTypes.value = emptyList()
        decodedPairPoints.value = emptyList()
        decodedPairIsGood.value = emptyList()
        decodedPairDescs.value = emptyList()
        decodedKalakini.value = com.miracle.names.data.model.KalakiniInfo()
        decodedProfile4D.value = null
        decodedMeaning.value = ""
        isDecodedFromDb.value = false
        _pinnedWishMeaning.value = ""
        _hasUserSelectedWishMeaning.value = false
        if (_randomMeaningSuggestions.value.isEmpty()) {
            _randomMeaningSuggestions.value = DEFAULT_AUSPICIOUS_MEANINGS.shuffled().take(3)
        }
        _isAutoSelectingMeaning.value = false
        _isMeaningAutoSelected.value = false
    }

    private fun fetchCelebrities() {
        _isCheckingConnection.value = true
        viewModelScope.launch {
            try {
                val response = RetrofitClient.apiService.getCelebrities()
                val body = response.body()
                if (!response.isSuccessful || body == null || body.error) {
                    _connectionError.value = CONNECTION_ERROR_MESSAGE
                    return@launch
                }

                _connectionError.value = null
                val data = body.data
                if (data.isNotEmpty()) {
                    celebrities.value = data
                    if (prototypeName.value.isNotEmpty()) {
                        decodeCurrentName(prototypeName.value)
                    }
                }
            } catch (_: Exception) {
                _connectionError.value = CONNECTION_ERROR_MESSAGE
            } finally {
                _isCheckingConnection.value = false
            }
        }
    }

    fun fetchArticles() {
        viewModelScope.launch {
            try {
                isLoadingArticles.value = true
                val response = RetrofitClient.apiService.getArticles()
                if (response.isSuccessful) {
                    val items = response.body()?.data.orEmpty()
                    if (items.isNotEmpty()) {
                        articles.value = items
                    }
                }
            } catch (_: Exception) {
                // Silently ignore failures for articles
            } finally {
                isLoadingArticles.value = false
            }
        }
    }

    fun retryBackendConnection() {
        fetchCelebrities()
        fetchArticles()
    }

    fun searchNames() {
        searchJob?.cancel()
        searchJob = null
        val searchRevision = searchRevisionGate.begin()
        invalidateRerankForSearch()
        _hasCompletedSearch.value = false

        val requestedMode = if (searchMode.value == MODE_COMBINED) MODE_COMBINED else MODE_STANDALONE
        val requestedFindSurnameMode = requestedMode == MODE_COMBINED && isFindSurnameMode.value
        val requestedSurname = resolveAppliedSurname(requestedMode, surname.value)
        val requestedDay = dayOfBirth.value.trim()
        val requestedWeights = _pendingRankingWeights.value
        val requestedPrototypeName = if (requestedMode == MODE_COMBINED) {
            requestedSurname
        } else {
            prototypeName.value.trim()
        }
        val currentSelectedWish = selectedWish.value.trim()

        val activePinnedMeaning = _pinnedWishMeaning.value.trim()
        val hasPinnedMeaning = _hasUserSelectedWishMeaning.value && activePinnedMeaning.isNotBlank()
        val finalWishQuery = when {
            currentSelectedWish.isNotEmpty() -> currentSelectedWish
            hasPinnedMeaning -> activePinnedMeaning
            else -> DEFAULT_WISH
        }

        if (requestedMode == MODE_COMBINED && requestedSurname.isBlank()) {
            if (searchRevisionGate.accepts(searchRevision)) {
                _uiState.value = UiState.Error(
                    if (requestedFindSurnameMode) "โปรดพิมพ์ชื่อตัวก่อนค้นหา"
                    else "โปรดพิมพ์นามสกุลก่อนค้นหา"
                )
            }
            return
        }

        if (requestedDay.isBlank()) {
            if (searchRevisionGate.accepts(searchRevision)) {
                _uiState.value = UiState.Error("โปรดเลือกวันเกิดเพื่อคัดกรองกาลกิณี")
            }
            return
        }

        if (!hasPinnedMeaning) {
            if (searchRevisionGate.accepts(searchRevision)) {
                _uiState.value = UiState.Error("โปรดเลือกความหมายมงคลก่อนค้นหา")
            }
            return
        }

        _uiState.value = UiState.Loading
        searchJob = viewModelScope.launch {
            try {
                val request = NameSearchRequest(
                    mode = requestedMode,
                    searchTarget = if (requestedFindSurnameMode) "surname" else "first_name",
                    surname = requestedSurname,
                    dayOfBirth = requestedDay,
                    prototypeName = requestedPrototypeName,
                    wishQuery = finalWishQuery,
                    numerologyWeight = requestedWeights.numerology,
                    semanticWeight = requestedWeights.semantic,
                    soundWeight = requestedWeights.sound,
                    similarityWeight = requestedWeights.similarity,
                    isNameIntent = requestedPrototypeName.isNotBlank(),
                    limit = 100
                )

                val response = RetrofitClient.apiService.searchTopNames(request)
                if (!searchRevisionGate.accepts(searchRevision)) return@launch
                val body = response.body()
                if (response.isSuccessful && body != null) {
                    if (!body.error && body.policy.rankingContractVersion == SUPPORTED_RANKING_CONTRACT) {
                        val appliedMode = resolveAppliedMode(
                            body.appliedMode.ifBlank { body.mode },
                            requestedMode
                        )
                        val responseRequestedWeights = resolveResponseWeights(
                            body.requestedWeights,
                            requestedWeights
                        )
                        val responseAppliedWeights = resolveResponseWeights(
                            body.appliedWeights,
                            responseRequestedWeights
                        )
                        val safeData = body.data ?: emptyList()
                        val resultCount = resolveActualCount(
                            body.returnedCount.takeIf { it > 0 || safeData.isEmpty() } ?: body.count,
                            safeData.size
                        )
                        val snapshot = AppliedSearchSnapshot(
                            mode = appliedMode,
                            surname = resolveAppliedSurname(appliedMode, requestedSurname),
                            isFindSurnameMode = requestedFindSurnameMode,
                            dayOfBirth = requestedDay,
                            wishQuery = finalWishQuery,
                            weights = responseAppliedWeights,
                            requestedWeights = responseRequestedWeights,
                            count = resultCount,
                            prototypeName = requestedPrototypeName,
                            eligibleCandidateCount = resolveEligibleCandidateCount(
                                body.eligibleCandidateCount,
                                resultCount
                            ),
                            semanticAvailable = body.semanticAvailable,
                            semanticSource = body.semanticSource,
                            semanticModel = body.resolvedSemanticQueryModel(),
                            semanticVectorHash = body.semanticVectorHash,
                            candidateVectorProvenanceVerified = body.candidateVectorProvenanceVerified,
                            candidateVectorProvenanceSource = body.candidateVectorProvenanceSource,
                            policy = body.policy,
                            appliedWeightsConfirmed = body.appliedWeights != null
                        )
                        candidateToken = body.candidateToken
                        _pendingRankingWeights.value = responseRequestedWeights
                        _uiState.value = UiState.Success(safeData, snapshot)
                        _hasCompletedSearch.value = true
                    } else {
                        _uiState.value = UiState.Error(
                            if (!body.error) {
                                "เวอร์ชันการจัดอันดับของระบบไม่ตรงกับแอป กรุณาอัปเดตแอปหรือค้นหาใหม่"
                            } else {
                                body.message
                            }
                        )
                    }
                } else {
                    _uiState.value = UiState.Error(
                        "API call failed: ${response.code()} ${response.message()}"
                    )
                }
            } catch (cancelled: CancellationException) {
                throw cancelled
            } catch (e: Exception) {
                if (searchRevisionGate.accepts(searchRevision)) {
                    _uiState.value = UiState.Error(
                        "Connection error: ${e.localizedMessage ?: "Unknown error"}"
                    )
                }
            } finally {
                if (searchRevisionGate.accepts(searchRevision)) {
                    searchJob = null
                }
            }
        }
    }

    private fun invalidateRerankForSearch() {
        rerankRequestRevision += 1
        rerankJob?.cancel()
        rerankJob = null
        movementClearJob?.cancel()
        movementClearJob = null
        _isReranking.value = false
        _rerankError.value = null
        rankMovements.value = emptyMap()
        newlyRankedIds.value = emptySet()
        rerankRevision.value = 0
        candidateToken = ""
        selectedItem.value = null
    }

    fun resetRankingWeightsToBalance() {
        _pendingRankingWeights.value = normalizeRankingWeights(0.25f, 0.25f, 0.25f, 0.25f)
    }

    fun selectNameForBlessingCard(item: NameResultItem?) {
        selectedItem.value = item
    }

    fun toggleGuideSheet(show: Boolean) {
        showGuideSheet.value = show
    }

    fun resetToForm() {
        val requestedWeights = (_uiState.value as? UiState.Success)?.snapshot?.requestedWeights
        searchRevisionGate.invalidate()
        searchJob?.cancel()
        searchJob = null
        rerankRequestRevision += 1
        rerankJob?.cancel()
        movementClearJob?.cancel()
        _isReranking.value = false
        _rerankError.value = null
        rankMovements.value = emptyMap()
        newlyRankedIds.value = emptySet()
        rerankRevision.value = 0
        candidateToken = ""
        if (requestedWeights != null) {
            _pendingRankingWeights.value = requestedWeights
        }
        _uiState.value = UiState.Idle
        selectedItem.value = null
    }

    fun resetAll() {
        searchRevisionGate.invalidate()
        searchJob?.cancel()
        searchJob = null
        rerankRequestRevision += 1
        rerankJob?.cancel()
        movementClearJob?.cancel()
        decodeJob?.cancel()
        suggestionJob?.cancel()

        prototypeName.value = ""
        surname.value = ""
        dayOfBirth.value = ""
        _nameSuggestions.value = emptyList()
        _suggestionPreviewPairs.value = emptyList()
        _suggestionCount.value = 0
        _suggestionsUnlocked.value = false
        _isLoadingNameSuggestions.value = false
        _selectedSuggestionName.value = ""
        selectedWish.value = DEFAULT_WISH
        _pinnedWishMeaning.value = ""
        _hasUserSelectedWishMeaning.value = false
        _randomMeaningSuggestions.value = DEFAULT_AUSPICIOUS_MEANINGS.shuffled().take(3)
        _isAutoSelectingMeaning.value = false
        _isMeaningAutoSelected.value = false

        decodedPairs.value = emptyList()
        decodedPairTypes.value = emptyList()
        decodedPairPoints.value = emptyList()
        decodedPairIsGood.value = emptyList()
        decodedPairDescs.value = emptyList()
        decodedKalakini.value = com.miracle.names.data.model.KalakiniInfo()
        decodedProfile4D.value = null
        decodedMeaning.value = ""
        isDecodedFromDb.value = false
        _isDecodingName.value = false
        _decodeError.value = null

        _isReranking.value = false
        _rerankError.value = null
        rankMovements.value = emptyMap()
        newlyRankedIds.value = emptySet()
        rerankRevision.value = 0
        candidateToken = ""
        _pendingRankingWeights.value = RankingWeights.Default
        _hasCompletedSearch.value = false
        _uiState.value = UiState.Idle
        selectedItem.value = null
    }
}

internal fun resolvedRank(item: NameResultItem, index: Int): Int {
    return item.rank.takeIf { it > 0 } ?: (index + 1)
}
