package com.miracle.names.service

import android.app.NotificationChannel
import android.app.NotificationManager
import android.app.PendingIntent
import android.content.Context
import android.content.Intent
import android.os.Build
import android.util.Log
import androidx.core.app.NotificationCompat
import com.google.firebase.messaging.FirebaseMessagingService
import com.google.firebase.messaging.RemoteMessage
import com.miracle.names.MainActivity
import com.miracle.names.R
import com.miracle.names.data.model.RegisterChatDeviceTokenRequest
import com.miracle.names.data.remote.RetrofitClient
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch

class MiracleFirebaseMessagingService : FirebaseMessagingService() {

    companion object {
        private const val TAG = "MiracleFCM"
        const val CHANNEL_ID = "chuedee_support_chat"
        const val CHANNEL_NAME = "ข้อความแชตปรึกษาชื่อมงคล"
    }

    override fun onNewToken(token: String) {
        super.onNewToken(token)
        Log.d(TAG, "New FCM token received: $token")
        val prefs = getSharedPreferences("support_chat", Context.MODE_PRIVATE)
        val conversationId = prefs.getString("conversation_id", "").orEmpty()
        val customerToken = prefs.getString("token", "").orEmpty()

        if (conversationId.isNotBlank() && customerToken.isNotBlank()) {
            CoroutineScope(Dispatchers.IO).launch {
                runCatching {
                    RetrofitClient.apiService.registerChatDeviceToken(
                        conversationId,
                        customerToken,
                        RegisterChatDeviceTokenRequest(
                            platform = "android",
                            deviceToken = token,
                            appVersion = "1.3.8"
                        )
                    )
                }.onSuccess {
                    Log.d(TAG, "Successfully registered new FCM token with backend")
                }.onFailure { e ->
                    Log.e(TAG, "Failed to register FCM token with backend", e)
                }
            }
        }
    }

    override fun onMessageReceived(remoteMessage: RemoteMessage) {
        super.onMessageReceived(remoteMessage)
        Log.d(TAG, "FCM Message received from: ${remoteMessage.from}")

        val title = remoteMessage.notification?.title 
            ?: remoteMessage.data["title"] 
            ?: "อ.ทญา ว่องไว"

        val body = remoteMessage.notification?.body 
            ?: remoteMessage.data["body"] 
            ?: "มีข้อความใหม่จากห้องแชตปรึกษา"

        val conversationId = remoteMessage.data["conversation_id"].orEmpty()

        showNotification(title, body, conversationId)
    }

    private fun showNotification(title: String, body: String, conversationId: String) {
        val notificationManager = getSystemService(Context.NOTIFICATION_SERVICE) as NotificationManager

        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            val channel = NotificationChannel(
                CHANNEL_ID,
                CHANNEL_NAME,
                NotificationManager.IMPORTANCE_HIGH
            ).apply {
                description = "การแจ้งเตือนข้อความตอบกลับจาก อ.ทญา ว่องไว"
                enableLights(true)
                enableVibration(true)
            }
            notificationManager.createNotificationChannel(channel)
        }

        val intent = Intent(this, MainActivity::class.java).apply {
            flags = Intent.FLAG_ACTIVITY_CLEAR_TOP or Intent.FLAG_ACTIVITY_SINGLE_TOP
            putExtra("open_chat", true)
            if (conversationId.isNotBlank()) {
                putExtra("conversation_id", conversationId)
            }
        }

        val pendingIntent = PendingIntent.getActivity(
            this,
            0,
            intent,
            PendingIntent.FLAG_UPDATE_CURRENT or PendingIntent.FLAG_IMMUTABLE
        )

        val largeLogo = android.graphics.BitmapFactory.decodeResource(resources, R.drawable.app_logo)

        val notification = NotificationCompat.Builder(this, CHANNEL_ID)
            .setSmallIcon(R.drawable.ic_notification_chuedee)
            .setColor(0xFFD4AF37.toInt())
            .setLargeIcon(largeLogo)
            .setContentTitle(title)
            .setContentText(body)
            .setStyle(NotificationCompat.BigTextStyle().bigText(body))
            .setPriority(NotificationCompat.PRIORITY_HIGH)
            .setAutoCancel(true)
            .setContentIntent(pendingIntent)
            .build()

        val notificationId = System.currentTimeMillis().toInt()
        notificationManager.notify(notificationId, notification)
    }
}
