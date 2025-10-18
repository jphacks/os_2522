package com.example.arsome.ml.face

import android.hardware.camera2.params.Face
import android.media.FaceDetector
import androidx.compose.ui.geometry.Rect
import com.google.mlkit.vision.face.Face

/**
 * 顔検出結果を表すデータクラス
 * Phase 2で認識情報を追加
 */
data class FaceDetectionResult(
    val face: Face,
    val boundingBox: Rect,
    val trackingId: Int?,
    val isStable: Boolean = false,
    val confidence: Float = 0f,
    val recognitionInfo: FaceRecognitionInfo? = null
) {
    /**
     * 顔の中心点を取得
     */
    val centerPoint: Pair<Float, Float>
        get() = Pair(boundingBox.center.x, boundingBox.center.y)
    
    /**
     * 顔のサイズ（短辺）を取得
     */
    val minSize: Float
        get() = minOf(boundingBox.width, boundingBox.height)
}

/**
 * セッション内での顔トラッキング情報
 */
data class FaceTrackingInfo(
    val trackingId: Int,
    var lastDetectionTime: Long = System.currentTimeMillis(),
    var lastCenterPoint: Pair<Float, Float>? = null,
    var stableStartTime: Long? = null,
    var isCurrentlyStable: Boolean = false,
    var recognizedPersonId: Long? = null,
    var unknownId: Int? = null,
    var cachedPersonInfo: CachedPersonInfo? = null, // Phase 4: 人物情報キャッシュ
    var hasAttemptedRecognition: Boolean = false
) {
    /**
     * 顔が安定しているかを判定
     */
    fun updateStability(
        currentCenter: Pair<Float, Float>,
        minSize: Float,
        currentTime: Long = System.currentTimeMillis()
    ): Boolean {
        val isValidSize = minSize >= com.example.arsome.constants.AppConstants.MIN_FACE_SIZE_PX
        val isValidMovement = lastCenterPoint?.let { lastCenter ->
            val distance = kotlin.math.sqrt(
                (currentCenter.first - lastCenter.first).let { it * it } +
                (currentCenter.second - lastCenter.second).let { it * it }
            )
            distance <= com.example.arsome.constants.AppConstants.MAX_FACE_CENTER_MOVEMENT_PX
        } ?: true
        
        val meetsStabilityCriteria = isValidSize && isValidMovement
        
        if (meetsStabilityCriteria) {
            if (stableStartTime == null) {
                stableStartTime = currentTime
            }
            val stableDuration = currentTime - (stableStartTime ?: currentTime)
            isCurrentlyStable = stableDuration >= com.example.arsome.constants.AppConstants.FACE_STABILITY_DURATION_MS
        } else {
            stableStartTime = null
            isCurrentlyStable = false
        }
        
        lastCenterPoint = currentCenter
        lastDetectionTime = currentTime
        
        return isCurrentlyStable
    }
}

/**
 * 顔認識情報を表すデータクラス
 */
data class FaceRecognitionInfo(
    val personId: Long?,
    val personName: String?,
    val unknownId: Int?,
    val confidence: Float = 0f,
    val lastSummary: String? = null
) {
    val displayName: String
        get() = personName ?: "Unknown #$unknownId"
    
    val isRecognized: Boolean
        get() = personId != null
}

/**
 * キャッシュされた人物情報
 */
data class CachedPersonInfo(
    val personName: String?,
    val lastSummary: String?
)
