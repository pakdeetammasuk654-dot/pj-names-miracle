package com.miracle.names.viewmodel

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.miracle.names.data.model.ChatMessage
import com.miracle.names.data.remote.RetrofitClient
import com.miracle.names.data.security.SecureChatTokenStore
import kotlinx.coroutines.Job
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch
import okhttp3.MediaType.Companion.toMediaTypeOrNull
import okhttp3.MultipartBody
import okhttp3.RequestBody.Companion.toRequestBody

object ChatAvatarHelper {
    val options = listOf(
        Triple("cat", "แมว", "/images/chat_avatar_cat.png"),
        Triple("dog", "สุนัข", "/images/chat_avatar_dog.png"),
        Triple("rabbit", "กระต่าย", "/images/chat_avatar_rabbit.png"),
        Triple("bear", "หมี", "/images/chat_avatar_bear.png"),
        Triple("panda", "แพนด้า", "/images/chat_avatar_panda.png")
    )

    fun toPath(keyOrPath: String): String {
        return when {
            keyOrPath.contains("cat") -> "/images/chat_avatar_cat.png"
            keyOrPath.contains("dog") -> "/images/chat_avatar_dog.png"
            keyOrPath.contains("rabbit") -> "/images/chat_avatar_rabbit.png"
            keyOrPath.contains("bear") -> "/images/chat_avatar_bear.png"
            keyOrPath.contains("panda") -> "/images/chat_avatar_panda.png"
            keyOrPath.contains("male") -> "/images/chat_avatar_dog.png"
            else -> "/images/chat_avatar_cat.png"
        }
    }

    fun toKey(keyOrPath: String): String {
        return when {
            keyOrPath.contains("cat") -> "cat"
            keyOrPath.contains("dog") -> "dog"
            keyOrPath.contains("rabbit") -> "rabbit"
            keyOrPath.contains("bear") -> "bear"
            keyOrPath.contains("panda") -> "panda"
            keyOrPath.contains("male") -> "dog"
            else -> "cat"
        }
    }

    fun toDrawableRes(keyOrPath: String): Int {
        return when {
            keyOrPath.contains("taya") -> com.miracle.names.R.drawable.taya_avatar
            keyOrPath.contains("app_logo") -> com.miracle.names.R.drawable.app_logo
            keyOrPath.contains("dog") || keyOrPath.contains("male") -> com.miracle.names.R.drawable.chat_avatar_dog
            keyOrPath.contains("rabbit") -> com.miracle.names.R.drawable.chat_avatar_rabbit
            keyOrPath.contains("bear") -> com.miracle.names.R.drawable.chat_avatar_bear
            keyOrPath.contains("panda") -> com.miracle.names.R.drawable.chat_avatar_panda
            else -> com.miracle.names.R.drawable.chat_avatar_cat
        }
    }
}

data class SupportChatState(
    val conversationId: String = "",
    val token: String = "",
    val messages: List<ChatMessage> = emptyList(),
    val unreadCount: Int = 0,
    val isLoading: Boolean = false,
    val isDeleting: Boolean = false,
    val isSending: Boolean = false,
    val error: String? = null,
    val customerAvatar: String = "/images/chat_avatar_cat.png"
) { val hasConversation: Boolean get() = conversationId.isNotBlank() && token.isNotBlank() }

class SupportChatViewModel(application: Application) : AndroidViewModel(application) {
    private val prefs = application.getSharedPreferences("support_chat", Application.MODE_PRIVATE)
    private val secureTokenStore = SecureChatTokenStore(application)
    private val initialToken = secureTokenStore.read().ifBlank {
        // One-time migration from the legacy plaintext preference.
        prefs.getString("token", "").orEmpty().also { legacy ->
            if (legacy.isNotBlank()) secureTokenStore.write(legacy)
            prefs.edit().remove("token").apply()
        }
    }
    private val _state = MutableStateFlow(
        SupportChatState(
            prefs.getString("conversation_id", "").orEmpty(),
            initialToken,
            customerAvatar = ChatAvatarHelper.toPath(prefs.getString("customer_avatar", "/images/chat_avatar_cat.png").orEmpty())
        )
    )
    val state: StateFlow<SupportChatState> = _state.asStateFlow()
    private var pollJob: Job? = null
    private var backgroundUnreadJob: Job? = null

    init {
        startBackgroundUnreadPolling()
    }

    fun open() {
        _state.value = _state.value.copy(unreadCount = 0)
        clearNotifications()
        if (_state.value.hasConversation) {
            refresh()
            startPolling()
            syncDeviceToken()
        }
    }

    fun clearNotifications() {
        val notificationManager = getApplication<Application>().getSystemService(android.content.Context.NOTIFICATION_SERVICE) as? android.app.NotificationManager
        notificationManager?.cancelAll()
    }

    fun close() {
        pollJob?.cancel()
        pollJob = null
        checkUnread()
    }

    fun selectAvatar(path: String) {
        val canonical = ChatAvatarHelper.toPath(path)
        prefs.edit().putString("customer_avatar", canonical).apply()
        _state.value = _state.value.copy(customerAvatar = canonical)
    }

    fun startNewConversation() {
        pollJob?.cancel()
        pollJob = null
        prefs.edit().remove("conversation_id").remove("token").apply()
        secureTokenStore.clear()
        _state.value = SupportChatState(customerAvatar = _state.value.customerAvatar)
    }

    fun deleteConversation() = viewModelScope.launch {
        val current = _state.value
        if (!current.hasConversation || current.isDeleting) return@launch
        _state.value = current.copy(isDeleting = true, error = null)
        runCatching {
            RetrofitClient.apiService.deleteChatConversation(current.conversationId, current.token)
        }.onSuccess { response ->
            if (response.isSuccessful) {
                startNewConversation()
            } else {
                _state.value = _state.value.copy(isDeleting = false, error = "ลบข้อมูลแชตไม่สำเร็จ กรุณาลองใหม่")
            }
        }.onFailure {
            _state.value = _state.value.copy(isDeleting = false, error = "ลบข้อมูลแชตไม่สำเร็จ กรุณาลองใหม่")
        }
    }

    fun startConversation(name: String) = viewModelScope.launch {
        _state.value = _state.value.copy(isLoading = true, error = null)
        val avatarKey = ChatAvatarHelper.toKey(_state.value.customerAvatar)
        runCatching {
            RetrofitClient.apiService.createChatConversation(
                name.trim().toRequestBody("text/plain".toMediaTypeOrNull()),
                avatarKey.toRequestBody("text/plain".toMediaTypeOrNull())
            )
        }.onSuccess { response ->
            val data = response.body()
            if (response.isSuccessful && data != null && data.conversationId.isNotBlank()) {
                val serverAvatar = ChatAvatarHelper.toPath(data.customerAvatar)
                prefs.edit().putString("conversation_id", data.conversationId).putString("customer_avatar", serverAvatar).remove("token").apply()
                secureTokenStore.write(data.customerToken)
                _state.value = _state.value.copy(
                    conversationId = data.conversationId,
                    token = data.customerToken,
                    customerAvatar = serverAvatar,
                    unreadCount = 0,
                    isLoading = false
                )
                refresh()
                startPolling()
                syncDeviceToken()
            } else {
                _state.value = _state.value.copy(isLoading = false, error = "เริ่มแชตไม่สำเร็จ")
            }
        }.onFailure {
            _state.value = _state.value.copy(isLoading = false, error = "เชื่อมต่อแชตไม่สำเร็จ")
        }
    }

    fun syncDeviceToken() {
        if (!_state.value.hasConversation) return
        runCatching {
            com.google.firebase.messaging.FirebaseMessaging.getInstance().token
                .addOnSuccessListener { token ->
                    if (!token.isNullOrBlank()) {
                        registerDeviceToken(token)
                    }
                }
        }
    }

    fun checkUnread() = viewModelScope.launch {
        val current = _state.value
        if (!current.hasConversation) {
            if (current.unreadCount != 0) _state.value = current.copy(unreadCount = 0)
            return@launch
        }
        runCatching {
            RetrofitClient.apiService.getChatUnreadCount(current.conversationId, current.token)
        }.onSuccess { response ->
            if (response.isSuccessful) {
                val count = response.body()?.unreadCount ?: 0
                _state.value = _state.value.copy(unreadCount = count)
            } else if (response.code() == 401) {
                startNewConversation()
            }
        }
    }

    fun refresh() = viewModelScope.launch {
        val current = _state.value
        if (!current.hasConversation) return@launch
        runCatching { RetrofitClient.apiService.getChatMessages(current.conversationId, current.token) }
            .onSuccess { response ->
                if (response.isSuccessful) {
                    val body = response.body()
                    val serverAvatar = body?.customerAvatar?.takeIf { it.isNotBlank() }?.let { ChatAvatarHelper.toPath(it) } ?: current.customerAvatar
                    prefs.edit().putString("customer_avatar", serverAvatar).apply()
                    _state.value = _state.value.copy(
                        messages = body?.messages.orEmpty(),
                        customerAvatar = serverAvatar,
                        unreadCount = 0,
                        error = null
                    )
                    clearNotifications()
                }
            }
    }

    fun send(body: String, imageBytes: ByteArray?, mediaType: String?) = viewModelScope.launch {
        val current = _state.value
        if (!current.hasConversation || (body.isBlank() && imageBytes == null)) return@launch
        _state.value = current.copy(isSending = true, error = null)
        val image = imageBytes?.let { MultipartBody.Part.createFormData("image", "question-image", it.toRequestBody((mediaType ?: "image/jpeg").toMediaTypeOrNull())) }
        runCatching { RetrofitClient.apiService.sendChatMessage(current.conversationId, current.token, body.trim().toRequestBody("text/plain".toMediaTypeOrNull()), image) }
            .onSuccess { response ->
                if (response.isSuccessful) {
                    _state.value = _state.value.copy(isSending = false, unreadCount = 0)
                    refresh()
                } else {
                    _state.value = _state.value.copy(isSending = false, error = "ส่งข้อความไม่สำเร็จ")
                }
            }
            .onFailure {
                _state.value = _state.value.copy(isSending = false, error = "ส่งข้อความไม่สำเร็จ")
            }
    }

    private fun startPolling() {
        pollJob?.cancel()
        pollJob = viewModelScope.launch {
            while (true) {
                delay(15_000)
                refresh()
            }
        }
    }

    private fun startBackgroundUnreadPolling() {
        backgroundUnreadJob?.cancel()
        backgroundUnreadJob = viewModelScope.launch {
            while (true) {
                if (pollJob == null || pollJob?.isActive != true) {
                    checkUnread()
                }
                // Push notifications are the primary background path. This
                // low-frequency fallback covers devices where push is delayed.
                delay(60_000)
            }
        }
    }

    fun registerDeviceToken(deviceToken: String) = viewModelScope.launch {
        val current = _state.value
        if (!current.hasConversation || deviceToken.isBlank()) return@launch
        runCatching {
            RetrofitClient.apiService.registerChatDeviceToken(
                current.conversationId,
                current.token,
                com.miracle.names.data.model.RegisterChatDeviceTokenRequest(
                    platform = "android",
                    deviceToken = deviceToken,
                    appVersion = "1.3.1"
                )
            )
        }
    }

    override fun onCleared() {
        pollJob?.cancel()
        backgroundUnreadJob?.cancel()
        super.onCleared()
    }
}
