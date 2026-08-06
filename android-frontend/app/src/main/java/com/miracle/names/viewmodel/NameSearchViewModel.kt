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

    private val _connectionError = MutableStateFlow<String?>(null)
    val connectionError: StateFlow<String?> = _connectionError.asStateFlow()

    private val _isCheckingConnection = MutableStateFlow(true)
    val isCheckingConnection: StateFlow<Boolean> = _isCheckingConnection.asStateFlow()

    // Form inputs state
    val searchMode = MutableStateFlow("standalone") // "standalone", "combined", "semantic"
    val surname = MutableStateFlow("")
    val dayOfBirth = MutableStateFlow("")
    val prototypeName = MutableStateFlow("")
    val selectedWish = MutableStateFlow(DEFAULT_WISH)
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
        const val CONNECTION_ERROR_MESSAGE = "ไม่สามารถเชื่อมต่อกับเซิร์ฟเวอร์ได้ กรุณาตรวจสอบการเชื่อมต่อแล้วลองใหม่"
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
                // The backend owns numerology calculations. Do not show a
                // different local result when the API is unavailable.
            }

            decodedPairs.value = emptyList()
            decodedPairTypes.value = emptyList()
            decodedMeaning.value = ""
            isDecodedFromDb.value = false
            _connectionError.value = CONNECTION_ERROR_MESSAGE
        }
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
            } catch (e: Exception) {
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
        viewModelScope.launch {
            _uiState.value = UiState.Loading
            try {
                val effectiveMode = searchMode.value
                val currentPrototype = prototypeName.value.trim()
                val currentSelectedWish = selectedWish.value.trim()

                // Determine meaning source
                val hasMeaningFromDb = isDecodedFromDb.value && decodedMeaning.value.isNotEmpty()
                val hasMeaningFromWish = currentSelectedWish.isNotEmpty()

                val (finalWishQuery, finalIsNameIntent) = when {
                    hasMeaningFromWish -> {
                        currentSelectedWish to false
                    }
                    hasMeaningFromDb -> {
                        // Use the name itself for vector similarity (name has embedding in DB)
                        currentPrototype to true
                    }
                    else -> {
                        // Fallback to default wish when no meaning in name or DB
                        DEFAULT_WISH to false
                    }
                }

                val request = NameSearchRequest(
                    mode = effectiveMode,
                    surname = surname.value,
                    dayOfBirth = dayOfBirth.value,
                    wishQuery = finalWishQuery,
                    semanticWeight = semanticWeight.value,
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
