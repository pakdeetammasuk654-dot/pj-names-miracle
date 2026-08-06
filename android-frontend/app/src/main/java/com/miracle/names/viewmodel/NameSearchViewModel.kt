package com.miracle.names.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.miracle.names.data.model.NameResultItem
import com.miracle.names.data.model.NameSearchRequest
import com.miracle.names.data.remote.RetrofitClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

sealed class UiState {
    object Idle : UiState()
    object Loading : UiState()
    data class Success(val results: List<NameResultItem>, val mode: String) : UiState()
    data class Error(val message: String) : UiState()
}

class NameSearchViewModel : ViewModel() {

    private val _uiState = MutableStateFlow<UiState>(UiState.Idle)
    val uiState: StateFlow<UiState> = _uiState.asStateFlow()

    // Form inputs state
    val searchMode = MutableStateFlow("standalone") // "standalone", "combined", "semantic"
    val surname = MutableStateFlow("")
    val dayOfBirth = MutableStateFlow("")
    val gender = MutableStateFlow("all") // "male", "female", "all"
    val taksaTarget = MutableStateFlow("sri")
    val prototypeName = MutableStateFlow("")
    val selectedWish = MutableStateFlow(DEFAULT_WISH)
    val filterTwoSyllables = MutableStateFlow(false)
    val semanticWeight = MutableStateFlow(0.5f)

    // Selected item for Blessing Card View
    val selectedItem = MutableStateFlow<NameResultItem?>(null)
    val showGuideSheet = MutableStateFlow(false)

    // Celebrities list
    val celebrities = MutableStateFlow<List<com.miracle.names.data.model.Celebrity>>(emptyList())

    // Decoded pairs from backend
    val decodedPairs = MutableStateFlow<List<String>>(emptyList())
    val decodedPairTypes = MutableStateFlow<List<String>>(emptyList())
    val decodedMeaning = MutableStateFlow("")
    val isDecodedFromDb = MutableStateFlow(false)

    companion object {
        const val DEFAULT_WISH = "🌸 เสริมดวงมงคลรอบด้าน (ทุกมิติชีวิต)"
    }

    init {
        fetchCelebrities()
    }

    fun decodeCurrentName(name: String) {
        if (name.isBlank()) {
            decodedPairs.value = emptyList()
            decodedPairTypes.value = emptyList()
            decodedMeaning.value = ""
            isDecodedFromDb.value = false
            return
        }

        viewModelScope.launch {
            try {
                val response = RetrofitClient.apiService.decodeName(name)
                if (response.isSuccessful && response.body() != null) {
                    val body = response.body()!!
                    decodedPairs.value = body.pairs ?: emptyList()
                    decodedPairTypes.value = body.pairTypes ?: emptyList()
                    decodedMeaning.value = body.meaning ?: ""
                    isDecodedFromDb.value = !body.meaning.isNullOrEmpty()
                    if (isDecodedFromDb.value) {
                        // Clear auto-selected wish — DB meaning takes priority
                        if (selectedWish.value == DEFAULT_WISH) {
                            selectedWish.value = ""
                        }
                    } else {
                        // Auto-select first preset wish when no meaning in DB (API success but no meaning)
                        if (selectedWish.value.isEmpty()) {
                            selectedWish.value = DEFAULT_WISH
                        }
                    }
                    return@launch
                }
            } catch (e: Exception) {
                // Fallback to local numerology calculation
            }
            
            // Local fallback
            val clean = name.trim().replace(" ", "")
            val numbers = com.miracle.names.utils.NumerologyCalculator.decodeString(clean)
            val pairs = com.miracle.names.utils.NumerologyCalculator.getPairs(numbers)
            val types = pairs.map { com.miracle.names.utils.NumerologyCalculator.getPairType(it) }
            decodedPairs.value = pairs
            decodedPairTypes.value = types
            decodedMeaning.value = ""
            isDecodedFromDb.value = false
            // Auto-select first preset wish when no meaning in DB
            if (selectedWish.value.isEmpty()) {
                selectedWish.value = DEFAULT_WISH
            }
        }
    }

    private fun fetchCelebrities() {
        viewModelScope.launch {
            try {
                val response = RetrofitClient.apiService.getCelebrities()
                if (response.isSuccessful) {
                    val data = response.body()?.data
                    if (!data.isNullOrEmpty()) {
                        celebrities.value = data
                        if (prototypeName.value.isEmpty()) {
                            val firstCeleb = data.first()
                            prototypeName.value = firstCeleb.name
                            decodeCurrentName(firstCeleb.name)
                        }
                    }
                }
            } catch (e: Exception) {
                // Ignore error, fallback to empty list or handled gracefully
            }
        }
    }

    fun searchNames() {
        viewModelScope.launch {
            _uiState.value = UiState.Loading
            try {
                val effectiveMode = searchMode.value
                val currentPrototype = prototypeName.value.trim()
                val currentSelectedWish = selectedWish.value.trim()

                // Determine meaning source
                val hasMeaningFromDb = isDecodedFromDb.value && decodedMeaning.value.isNotEmpty()
                val hasMeaningFromWish = currentSelectedWish.isNotEmpty()

                var finalWishQuery = ""
                var finalIsNameIntent = false
                var effectiveSemanticWeight = semanticWeight.value

                when {
                    hasMeaningFromWish -> {
                        finalWishQuery = currentSelectedWish
                        finalIsNameIntent = false
                    }
                    hasMeaningFromDb -> {
                        // Use the name itself for vector similarity (name has embedding in DB)
                        finalWishQuery = currentPrototype
                        finalIsNameIntent = true
                    }
                    else -> {
                        // Fallback to default wish when no meaning in name or DB
                        finalWishQuery = DEFAULT_WISH
                        finalIsNameIntent = false
                    }
                }

                val request = NameSearchRequest(
                    mode = effectiveMode,
                    surname = surname.value,
                    dayOfBirth = dayOfBirth.value,
                    gender = gender.value,
                    taksaTarget = taksaTarget.value,
                    wishQuery = finalWishQuery,
                    minLength = if (filterTwoSyllables.value) 4 else 0,
                    semanticWeight = effectiveSemanticWeight,
                    isNameIntent = finalIsNameIntent,
                    limit = 100
                )

                val response = RetrofitClient.apiService.searchTopNames(request)
                if (response.isSuccessful && response.body() != null) {
                    val body = response.body()!!
                    if (!body.error) {
                        _uiState.value = UiState.Success(body.data, body.mode)
                    } else {
                        _uiState.value = UiState.Error(body.message)
                    }
                } else {
                    _uiState.value = UiState.Error("API call failed: ${response.code()} ${response.message()}")
                }
            } catch (e: Exception) {
                _uiState.value = UiState.Error("Connection error: ${e.localizedMessage ?: "Unknown error"}")
            }
        }
    }

    fun selectNameForBlessingCard(item: NameResultItem?) {
        selectedItem.value = item
    }

    fun toggleGuideSheet(show: Boolean) {
        showGuideSheet.value = show
    }

    fun resetToForm() {
        _uiState.value = UiState.Idle
        selectedItem.value = null
    }
}
