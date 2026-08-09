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
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
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

class NameSearchViewModel : ViewModel() {

    private val _uiState = MutableStateFlow<UiState>(UiState.Idle)
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    private val _connectionError = MutableStateFlow<String?>(null)
    val connectionError: StateFlow<String?> = _connectionError.asStateFlow()

    private val _isCheckingConnection = MutableStateFlow(true)
    val isCheckingConnection: StateFlow<Boolean> = _isCheckingConnection.asStateFlow()

    val searchMode = MutableStateFlow(MODE_STANDALONE)
    val isFindSurnameMode = MutableStateFlow(false)
    val surname = MutableStateFlow("")
    val dayOfBirth = MutableStateFlow("")
    val prototypeName = MutableStateFlow("")
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
    val rerankRevision = MutableStateFlow(0)
    private var rerankJob: Job? = null
    private var movementClearJob: Job? = null
    private var rerankRequestRevision = 0L

    // Selected item for Blessing Card View.
    val selectedItem = MutableStateFlow<NameResultItem?>(null)
    val showGuideSheet = MutableStateFlow(false)

    // Celebrities list.
    val celebrities = MutableStateFlow<List<com.miracle.names.data.model.Celebrity>>(emptyList())

    // Decoded pairs from backend.
    val decodedPairs = MutableStateFlow<List<String>>(emptyList())
    val decodedPairTypes = MutableStateFlow<List<String>>(emptyList())
    val decodedPairPoints = MutableStateFlow<List<Int>>(emptyList())
    val decodedPairIsGood = MutableStateFlow<List<Boolean>>(emptyList())
    val decodedMeaning = MutableStateFlow("")
    val isDecodedFromDb = MutableStateFlow(false)
    private val _isDecodingName = MutableStateFlow(false)
    val isDecodingName: StateFlow<Boolean> = _isDecodingName.asStateFlow()
    private val _decodeError = MutableStateFlow<String?>(null)
    val decodeError: StateFlow<String?> = _decodeError.asStateFlow()
    private var decodeJob: Job? = null
    private var decodeRequestRevision = 0L

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
    }

    init {
        fetchCelebrities()
    }

    fun updateSearchMode(mode: String) {
        val resolvedMode = if (mode == MODE_COMBINED) MODE_COMBINED else MODE_STANDALONE
        searchMode.value = resolvedMode
        if (resolvedMode == MODE_STANDALONE) {
            // A hidden surname must never leak into a standalone result or its detail card.
            surname.value = ""
            decodeCurrentName(prototypeName.value)
        } else {
            decodeCurrentName(surname.value)
        }
    }

    fun updateRankingWeights(numerology: Float, semantic: Float, sound: Float) {
        _pendingRankingWeights.value = normalizeRankingWeights(numerology, semantic, sound)
    }

    fun updateRankingWeightsAndRerank(numerology: Float, semantic: Float, sound: Float) {
        val requestedWeights = normalizeRankingWeights(numerology, semantic, sound)
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
                        limit = 100
                    )
                )
                if (requestRevision != rerankRequestRevision) return@launch

                val body = response.body()
                if (response.isSuccessful && body != null && !body.error) {
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
                    val resultCount = resolveActualCount(
                        body.returnedCount.takeIf { it > 0 || body.data.isEmpty() } ?: body.count,
                        body.data.size
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
                    }
                } else {
                    val message = body?.message?.takeIf { it.isNotBlank() }
                        ?: "ไม่สามารถปรับอันดับได้ กรุณาลองใหม่"
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
                val response = RetrofitClient.apiService.decodeName(requestedName)
                if (!isCurrentDecodeRequest(requestRevision, requestedName)) return@launch

                val body = response.body()
                if (response.isSuccessful && body != null && !body.error) {
                    decodedPairs.value = body.pairs.orEmpty()
                    decodedPairTypes.value = body.pairTypes.orEmpty()
                    decodedPairPoints.value = body.pairPoints.orEmpty()
                    decodedPairIsGood.value = body.pairIsGood.orEmpty()
                    decodedMeaning.value = body.meaning.orEmpty()
                    isDecodedFromDb.value = !body.meaning.isNullOrBlank()
                    _connectionError.value = null

                    if (isDecodedFromDb.value) {
                        if (selectedWish.value == DEFAULT_WISH || selectedWish.value.startsWith("💡")) {
                            selectedWish.value = ""
                        }
                    } else {
                        selectedWish.value = DEFAULT_WISH
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

    private fun isCurrentDecodeRequest(requestRevision: Long, requestedName: String): Boolean {
        val currentTarget = if (searchMode.value == MODE_COMBINED) surname.value else prototypeName.value
        return requestRevision == decodeRequestRevision && currentTarget == requestedName
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
        decodedMeaning.value = ""
        isDecodedFromDb.value = false
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
                    if (prototypeName.value.isEmpty()) {
                        val firstCeleb = data.first()
                        prototypeName.value = firstCeleb.name
                        decodeCurrentName(firstCeleb.name)
                    } else {
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

    fun retryBackendConnection() {
        fetchCelebrities()
    }

    fun searchNames() {
        searchJob?.cancel()
        searchJob = null
        val searchRevision = searchRevisionGate.begin()
        invalidateRerankForSearch()

        val requestedMode = if (searchMode.value == MODE_COMBINED) MODE_COMBINED else MODE_STANDALONE
        val requestedSurname = resolveAppliedSurname(requestedMode, surname.value)
        val requestedDay = dayOfBirth.value.trim()
        val requestedWeights = _pendingRankingWeights.value
        val currentSelectedWish = selectedWish.value.trim()

        val hasMeaningFromDb = isDecodedFromDb.value && decodedMeaning.value.isNotBlank()
        val finalWishQuery = when {
            currentSelectedWish.isNotEmpty() -> currentSelectedWish
            hasMeaningFromDb -> decodedMeaning.value.trim()
            else -> DEFAULT_WISH
        }

        if (requestedMode == MODE_COMBINED && requestedSurname.isBlank()) {
            if (searchRevisionGate.accepts(searchRevision)) {
                _uiState.value = UiState.Error("โปรดกรอกนามสกุลก่อนค้นหา")
            }
            return
        }

        _uiState.value = UiState.Loading
        searchJob = viewModelScope.launch {
            try {
                val request = NameSearchRequest(
                    mode = requestedMode,
                    surname = requestedSurname,
                    dayOfBirth = requestedDay,
                    wishQuery = finalWishQuery,
                    numerologyWeight = requestedWeights.numerology,
                    semanticWeight = requestedWeights.semantic,
                    soundWeight = requestedWeights.sound,
                    isNameIntent = false,
                    limit = 100
                )

                val response = RetrofitClient.apiService.searchTopNames(request)
                if (!searchRevisionGate.accepts(searchRevision)) return@launch
                val body = response.body()
                if (response.isSuccessful && body != null) {
                    if (!body.error) {
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
                        val resultCount = resolveActualCount(
                            body.returnedCount.takeIf { it > 0 || body.data.isEmpty() } ?: body.count,
                            body.data.size
                        )
                        val snapshot = AppliedSearchSnapshot(
                            mode = appliedMode,
                            surname = resolveAppliedSurname(appliedMode, requestedSurname),
                            isFindSurnameMode = isFindSurnameMode.value,
                            dayOfBirth = requestedDay,
                            wishQuery = finalWishQuery,
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
                        candidateToken = body.candidateToken
                        _pendingRankingWeights.value = responseRequestedWeights
                        _uiState.value = UiState.Success(body.data, snapshot)
                    } else {
                        _uiState.value = UiState.Error(body.message)
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
        rerankRevision.value = 0
        candidateToken = ""
        selectedItem.value = null
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
        rerankRevision.value = 0
        candidateToken = ""
        if (requestedWeights != null) {
            _pendingRankingWeights.value = requestedWeights
        }
        _uiState.value = UiState.Idle
        selectedItem.value = null
    }
}

internal fun resolvedRank(item: NameResultItem, index: Int): Int {
    return item.rank.takeIf { it > 0 } ?: (index + 1)
}
