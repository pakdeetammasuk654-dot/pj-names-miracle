package com.miracle.names.viewmodel

import androidx.lifecycle.ViewModel
import androidx.lifecycle.viewModelScope
import com.google.gson.JsonParser
import com.miracle.names.data.model.AuspiciousActivity
import com.miracle.names.data.model.AuspiciousSearchRequest
import com.miracle.names.data.model.AuspiciousSearchResponse
import com.miracle.names.data.model.TaksaCalculationResponse
import com.miracle.names.data.remote.RetrofitClient
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.flow.update
import kotlinx.coroutines.launch
import kotlinx.coroutines.Job

data class AuspiciousUiState(
    val activities: List<AuspiciousActivity> = emptyList(),
    val selectedActivityCode: String = "",
    val isLoadingActivities: Boolean = true,
    val birthDate: String = "",
    val birthTime: String = "",
    val isSearching: Boolean = false,
    val isLoadingTaksa: Boolean = false,
    val taksa: TaksaCalculationResponse? = null,
    val taksaErrorMessage: String? = null,
    val result: AuspiciousSearchResponse? = null,
    val errorMessage: String? = null
) {
    val selectedActivity: AuspiciousActivity?
        get() = activities.firstOrNull { it.code == selectedActivityCode }
}

class AuspiciousViewModel : ViewModel() {
    private val api = RetrofitClient.apiService
    private val _state = MutableStateFlow(AuspiciousUiState())
    val state: StateFlow<AuspiciousUiState> = _state.asStateFlow()
    private var taksaJob: Job? = null

    init {
        loadActivities()
    }

    fun loadActivities() {
        viewModelScope.launch {
            _state.update { it.copy(isLoadingActivities = true, errorMessage = null) }
            try {
                val response = api.getAuspiciousActivities()
                if (!response.isSuccessful) {
                    throw IllegalStateException(response.errorMessage("ไม่สามารถโหลดรายการฤกษ์ได้"))
                }
                val activities = response.body()?.activities.orEmpty()
                if (activities.isEmpty()) throw IllegalStateException("ยังไม่พบกิจกรรมสำหรับดูฤกษ์")
                val preferred = activities.firstOrNull { it.code == "name_change" } ?: activities.first()
                _state.update {
                    it.copy(
                        activities = activities,
                        selectedActivityCode = if (it.selectedActivityCode.isBlank()) preferred.code else it.selectedActivityCode,
                        isLoadingActivities = false
                    )
                }
            } catch (error: Exception) {
                _state.update {
                    it.copy(
                        isLoadingActivities = false,
                        errorMessage = error.message ?: "เชื่อมต่อระบบดูฤกษ์ไม่สำเร็จ"
                    )
                }
            }
        }
    }

    fun selectActivity(code: String) {
        _state.update { it.copy(selectedActivityCode = code, result = null, errorMessage = null) }
    }

    fun updateBirthDate(value: String) {
        _state.update {
            it.copy(
                birthDate = value,
                birthTime = if (value.isBlank()) "" else it.birthTime
            )
        }
    }

    fun updateBirthTime(value: String) {
        _state.update { it.copy(birthTime = value) }
    }

    fun search(request: AuspiciousSearchRequest) {
        viewModelScope.launch {
            _state.update { it.copy(isSearching = true, result = null, errorMessage = null) }
            try {
                val response = api.searchAuspiciousDates(request)
                if (!response.isSuccessful) {
                    throw IllegalStateException(response.errorMessage("ค้นหาฤกษ์ไม่สำเร็จ"))
                }
                _state.update { it.copy(isSearching = false, result = response.body()) }
            } catch (error: Exception) {
                _state.update {
                    it.copy(
                        isSearching = false,
                        errorMessage = error.message ?: "เชื่อมต่อระบบดูฤกษ์ไม่สำเร็จ"
                    )
                }
            }
        }
    }

    fun loadTaksa(birthDate: String, birthTime: String) {
        taksaJob?.cancel()
        taksaJob = viewModelScope.launch {
            _state.update { it.copy(isLoadingTaksa = true, taksaErrorMessage = null) }
            try {
                val response = api.getTaksaCalculation(
                    birthDate = birthDate,
                    birthTime = birthTime
                )
                if (!response.isSuccessful) {
                    throw IllegalStateException(response.errorMessage("คำนวณทักษาจรไม่สำเร็จ"))
                }
                _state.update {
                    it.copy(
                        isLoadingTaksa = false,
                        taksa = response.body(),
                        taksaErrorMessage = null
                    )
                }
            } catch (error: Exception) {
                if (error is kotlinx.coroutines.CancellationException) throw error
                _state.update {
                    it.copy(
                        isLoadingTaksa = false,
                        taksaErrorMessage = error.message ?: "เชื่อมต่อระบบทักษาจรไม่สำเร็จ"
                    )
                }
            }
        }
    }

    fun clearTaksa() {
        taksaJob?.cancel()
        _state.update { it.copy(taksa = null, isLoadingTaksa = false, taksaErrorMessage = null) }
    }

    fun clearError() {
        _state.update { it.copy(errorMessage = null) }
    }
}

private fun <T> retrofit2.Response<T>.errorMessage(fallback: String): String {
    val raw = errorBody()?.string().orEmpty()
    val parsed: String? = runCatching {
        JsonParser().parse(raw).asJsonObject.get("error")?.asString
    }.getOrNull()
    return parsed?.takeIf { it.isNotBlank() } ?: fallback
}
