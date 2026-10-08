package com.miracle.names.utils

import android.content.ContentValues
import android.content.Context
import android.graphics.Bitmap
import android.net.Uri
import android.os.Build
import android.os.Environment
import android.provider.MediaStore
import java.io.OutputStream

object ImageUtils {
    fun saveBitmapToGallery(context: Context, bitmap: Bitmap): Boolean {
        val filename = "MiracleBlessing_${System.currentTimeMillis()}.jpg"
        var fos: OutputStream? = null
        var imageUri: Uri? = null

        try {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                val resolver = context.contentResolver
                val contentValues = ContentValues().apply {
                    put(MediaStore.MediaColumns.DISPLAY_NAME, filename)
                    put(MediaStore.MediaColumns.MIME_TYPE, "image/jpeg")
                    put(MediaStore.MediaColumns.RELATIVE_PATH, Environment.DIRECTORY_PICTURES + "/MiracleNames")
                }
                imageUri = resolver.insert(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, contentValues)
                if (imageUri != null) {
                    fos = resolver.openOutputStream(imageUri)
                }
            } else {
                // Fallback for older devices (Requires WRITE_EXTERNAL_STORAGE)
                val imagesDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_PICTURES).toString() + "/MiracleNames"
                val fileDir = java.io.File(imagesDir)
                if (!fileDir.exists()) {
                    fileDir.mkdirs()
                }
                val imageFile = java.io.File(fileDir, filename)
                fos = java.io.FileOutputStream(imageFile)
                
                // Add to gallery
                val contentValues = ContentValues().apply {
                    put(MediaStore.Images.Media.DATA, imageFile.absolutePath)
                    put(MediaStore.Images.Media.MIME_TYPE, "image/jpeg")
                }
                context.contentResolver.insert(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, contentValues)
            }

            fos?.use {
                bitmap.compress(Bitmap.CompressFormat.JPEG, 100, it)
                return true
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
        return false
    }

    fun saveBitmapAndGetUri(context: Context, bitmap: Bitmap): Uri? {
        val filename = "MiracleBlessing_${System.currentTimeMillis()}.jpg"
        var fos: OutputStream? = null
        var imageUri: Uri? = null

        try {
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.Q) {
                val resolver = context.contentResolver
                val contentValues = ContentValues().apply {
                    put(MediaStore.MediaColumns.DISPLAY_NAME, filename)
                    put(MediaStore.MediaColumns.MIME_TYPE, "image/jpeg")
                    put(MediaStore.MediaColumns.RELATIVE_PATH, Environment.DIRECTORY_PICTURES + "/MiracleNames")
                }
                imageUri = resolver.insert(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, contentValues)
                if (imageUri != null) {
                    fos = resolver.openOutputStream(imageUri)
                }
            } else {
                val imagesDir = Environment.getExternalStoragePublicDirectory(Environment.DIRECTORY_PICTURES).toString() + "/MiracleNames"
                val fileDir = java.io.File(imagesDir)
                if (!fileDir.exists()) {
                    fileDir.mkdirs()
                }
                val imageFile = java.io.File(fileDir, filename)
                fos = java.io.FileOutputStream(imageFile)
                val contentValues = ContentValues().apply {
                    put(MediaStore.Images.Media.DATA, imageFile.absolutePath)
                    put(MediaStore.Images.Media.MIME_TYPE, "image/jpeg")
                }
                imageUri = context.contentResolver.insert(MediaStore.Images.Media.EXTERNAL_CONTENT_URI, contentValues)
            }

            fos?.use {
                bitmap.compress(Bitmap.CompressFormat.JPEG, 100, it)
                return imageUri
            }
        } catch (e: Exception) {
            e.printStackTrace()
        }
        return null
    }

    fun shareImage(context: Context, bitmap: Bitmap) {
        try {
            val cachePath = java.io.File(context.cacheDir, "shared_images")
            cachePath.mkdirs()
            val file = java.io.File(cachePath, "miracle_blessing_${System.currentTimeMillis()}.jpg")
            val stream = java.io.FileOutputStream(file)
            bitmap.compress(Bitmap.CompressFormat.JPEG, 100, stream)
            stream.close()

            val contentUri = androidx.core.content.FileProvider.getUriForFile(
                context,
                "${context.packageName}.fileprovider",
                file
            )

            if (contentUri != null) {
                val shareIntent = android.content.Intent(android.content.Intent.ACTION_SEND).apply {
                    type = "image/jpeg"
                    putExtra(android.content.Intent.EXTRA_STREAM, contentUri)
                    addFlags(android.content.Intent.FLAG_GRANT_READ_URI_PERMISSION)
                }

                val chooser = android.content.Intent.createChooser(shareIntent, "แชร์การ์ดชื่อดีมงคลไปยัง")
                
                val resInfoList = context.packageManager.queryIntentActivities(chooser, android.content.pm.PackageManager.MATCH_DEFAULT_ONLY)
                for (resolveInfo in resInfoList) {
                    val packageName = resolveInfo.activityInfo.packageName
                    context.grantUriPermission(packageName, contentUri, android.content.Intent.FLAG_GRANT_READ_URI_PERMISSION)
                }

                context.startActivity(chooser)
            }
        } catch (e: Exception) {
            e.printStackTrace()
            // Fallback for older devices/edge cases
            val uri = saveBitmapAndGetUri(context, bitmap) ?: return
            val shareIntent = android.content.Intent(android.content.Intent.ACTION_SEND).apply {
                type = "image/jpeg"
                putExtra(android.content.Intent.EXTRA_STREAM, uri)
                addFlags(android.content.Intent.FLAG_GRANT_READ_URI_PERMISSION)
            }
            context.startActivity(android.content.Intent.createChooser(shareIntent, "แชร์การ์ดชื่อดีมงคลไปยัง"))
        }
    }

    fun createWallpaperBitmap(context: Context, cardBitmap: Bitmap): Bitmap {
        val displayMetrics = context.resources.displayMetrics
        val screenWidth = displayMetrics.widthPixels.coerceAtLeast(1080)
        val screenHeight = displayMetrics.heightPixels.coerceAtLeast(2280)

        val wallpaperBitmap = Bitmap.createBitmap(screenWidth, screenHeight, Bitmap.Config.ARGB_8888)
        val canvas = android.graphics.Canvas(wallpaperBitmap)

        // 1. Draw full-bleed vertical background gradient matching phone screen height
        val bgPaint = android.graphics.Paint(android.graphics.Paint.ANTI_ALIAS_FLAG).apply {
            shader = android.graphics.LinearGradient(
                0f, 0f, 0f, screenHeight.toFloat(),
                intArrayOf(
                    0xFFFFFDEB.toInt(), // Soft Cream Gold (Top safe area for clock)
                    0xFFFEF08A.toInt(), // Radiant Golden Aura
                    0xFFFDE047.toInt(), // Pure Amber Gold
                    0xFFF59E0B.toInt(), // Luminous Warm Gold
                    0xFFD97706.toInt(), // Rich Imperial Gold
                    0xFFB45309.toInt()  // Deep Amber Gold (Bottom safe area for app dock)
                ),
                null,
                android.graphics.Shader.TileMode.CLAMP
            )
        }
        canvas.drawRect(0f, 0f, screenWidth.toFloat(), screenHeight.toFloat(), bgPaint)

        // 2. Position the captured Blessing Card in safe center area
        val cardAspect = cardBitmap.width.toFloat() / cardBitmap.height.toFloat()
        var targetWidth = (screenWidth * 0.88f).toInt()
        var targetHeight = (targetWidth / cardAspect).toInt()

        val maxHeightAllowed = (screenHeight * 0.70f).toInt()
        if (targetHeight > maxHeightAllowed) {
            targetHeight = maxHeightAllowed
            targetWidth = (targetHeight * cardAspect).toInt()
        }

        val left = (screenWidth - targetWidth) / 2
        // Place slightly above center to avoid bottom app dock icons
        val top = ((screenHeight - targetHeight) * 0.40f).toInt()
        val right = left + targetWidth
        val bottom = top + targetHeight

        val destRect = android.graphics.Rect(left, top, right, bottom)
        val srcRect = android.graphics.Rect(0, 0, cardBitmap.width, cardBitmap.height)

        val cardPaint = android.graphics.Paint(android.graphics.Paint.ANTI_ALIAS_FLAG or android.graphics.Paint.FILTER_BITMAP_FLAG)
        canvas.drawBitmap(cardBitmap, srcRect, destRect, cardPaint)

        return wallpaperBitmap
    }
}
