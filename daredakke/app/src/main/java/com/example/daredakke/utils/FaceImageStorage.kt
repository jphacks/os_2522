package com.example.daredakke.utils

import android.content.Context
import android.graphics.Bitmap
import java.io.File
import java.io.FileOutputStream

/**
 * 顔サムネイル画像の保存・管理ユーティリティ
 */
object FaceImageStorage {
    
    private const val FACE_IMAGES_DIR = "face_images"
    
    /**
     * 指定した人物のサムネイルを保存し、保存先パスを返す
     */
    fun saveFaceThumbnail(
        context: Context,
        personId: Long,
        bitmap: Bitmap
    ): String? {
        return try {
            val personDir = File(
                context.filesDir,
                "$FACE_IMAGES_DIR/$personId"
            ).apply {
                if (!exists()) {
                    mkdirs()
                } else {
                    // 代表画像は最新1枚に絞る
                    listFiles()?.forEach { existing ->
                        if (existing.isFile) {
                            existing.delete()
                        }
                    }
                }
            }
            
            val fileName = "face_${System.currentTimeMillis()}.jpg"
            val outputFile = File(personDir, fileName)
            
            FileOutputStream(outputFile).use { outputStream ->
                bitmap.compress(Bitmap.CompressFormat.JPEG, 90, outputStream)
            }
            
            outputFile.absolutePath
        } catch (e: Exception) {
            println("Failed to save face thumbnail: ${e.message}")
            null
        }
    }
}
