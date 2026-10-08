package com.miracle.names.data.security

import android.content.Context
import android.security.keystore.KeyGenParameterSpec
import android.security.keystore.KeyProperties
import android.util.Base64
import java.security.KeyStore
import javax.crypto.Cipher
import javax.crypto.KeyGenerator
import javax.crypto.SecretKey
import javax.crypto.spec.GCMParameterSpec

object EntitlementTokenProvider {
    @Volatile var token: String = ""
        private set

    fun update(value: String) {
        token = value.trim()
    }
}

class SecureEntitlementStore(context: Context) {
    private val prefs = context.getSharedPreferences("premium_entitlement_secure", Context.MODE_PRIVATE)
    private val keyStore = KeyStore.getInstance("AndroidKeyStore").apply { load(null) }

    fun load(): String = runCatching {
        val encrypted = Base64.decode(prefs.getString("token", ""), Base64.NO_WRAP)
        val iv = Base64.decode(prefs.getString("iv", ""), Base64.NO_WRAP)
        if (encrypted.isEmpty() || iv.isEmpty()) return ""
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.DECRYPT_MODE, key(), GCMParameterSpec(128, iv))
        String(cipher.doFinal(encrypted), Charsets.UTF_8)
    }.getOrDefault("")

    fun save(token: String) {
        val cipher = Cipher.getInstance(TRANSFORMATION)
        cipher.init(Cipher.ENCRYPT_MODE, key())
        prefs.edit()
            .putString("token", Base64.encodeToString(cipher.doFinal(token.toByteArray()), Base64.NO_WRAP))
            .putString("iv", Base64.encodeToString(cipher.iv, Base64.NO_WRAP))
            .apply()
        EntitlementTokenProvider.update(token)
    }

    fun clear() {
        prefs.edit().clear().apply()
        EntitlementTokenProvider.update("")
    }

    private fun key(): SecretKey {
        (keyStore.getKey(KEY_ALIAS, null) as? SecretKey)?.let { return it }
        return KeyGenerator.getInstance(KeyProperties.KEY_ALGORITHM_AES, "AndroidKeyStore").run {
            init(
                KeyGenParameterSpec.Builder(
                    KEY_ALIAS,
                    KeyProperties.PURPOSE_ENCRYPT or KeyProperties.PURPOSE_DECRYPT
                ).setBlockModes(KeyProperties.BLOCK_MODE_GCM)
                    .setEncryptionPaddings(KeyProperties.ENCRYPTION_PADDING_NONE)
                    .build()
            )
            generateKey()
        }
    }

    private companion object {
        const val KEY_ALIAS = "chuedee_premium_entitlement_v1"
        const val TRANSFORMATION = "AES/GCM/NoPadding"
    }
}
