package com.example.daredakke.ml.face

import android.content.Context
import android.graphics.Bitmap
import androidx.camera.core.ExperimentalGetImage
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.ImageProxy
import com.google.mlkit.vision.common.InputImage
import com.google.mlkit.vision.face.FaceDetection
import com.google.mlkit.vision.face.FaceDetectorOptions
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import androidx.compose.ui.geometry.Rect
import com.example.daredakke.constants.AppConstants
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.launch
import kotlinx.coroutines.withContext
import com.example.daredakke.data.repository.PersonRepository
import com.example.daredakke.utils.FaceImageStorage
import kotlin.math.roundToInt

/**
 * ML Kit を使用した顔検出クラス
 * Phase 1: リアルタイム顔検出とセッション内トラッキング
 * Phase 2: 顔埋め込み抽出と認識機能を統合
 * Phase 4: 再会時の要約表示機能
 */
class FaceDetector(
    private val appContext: Context,
    private val embeddingExtractor: FaceEmbeddingExtractor? = null,
    private val faceRecognizer: FaceRecognizer? = null,
    private val personRepository: PersonRepository? = null
) : ImageAnalysis.Analyzer {
    
    private val faceDetectorOptions = FaceDetectorOptions.Builder()
        .setPerformanceMode(FaceDetectorOptions.PERFORMANCE_MODE_FAST)
        .setLandmarkMode(FaceDetectorOptions.LANDMARK_MODE_ALL)
        .setClassificationMode(FaceDetectorOptions.CLASSIFICATION_MODE_NONE)
        .setMinFaceSize(0.1f) // 画像の10%以上のサイズの顔を検出
        .enableTracking()
        .build()
    
    private val detector = FaceDetection.getClient(faceDetectorOptions)
    
    private val _detectionResults = MutableStateFlow<List<FaceDetectionResult>>(emptyList())
    val detectionResults: StateFlow<List<FaceDetectionResult>> = _detectionResults.asStateFlow()
    
    // セッション内でのトラッキング情報を保持
    private val trackingInfoMap = mutableMapOf<Int, FaceTrackingInfo>()
    private var unknownIdCounter = 1
    
    // 埋め込み抽出の間隔制御
    private val lastEmbeddingExtractionTime = mutableMapOf<Int, Long>()
    
    // 最後に抽出した埋め込みを保存
    private val lastExtractedEmbeddings = mutableMapOf<Int, FloatArray>()
    
    // 最後に切り出した顔サムネイルを保存
    private val lastCapturedFaceBitmaps = mutableMapOf<Int, Bitmap>()
    
    
    // 座標変換用のプレビューサイズ
    private var previewWidth: Int = 0
    private var previewHeight: Int = 0
    private var isUsingFrontCamera: Boolean = true
    
    // デバッグ用：座標変換モード
    private enum class CoordinateTransformMode {
        RAW,        // 生座標（変換なし）
        SIMPLE,     // シンプルなスケール変換
        COMPLEX     // 複雑な変換（回転・反転対応）
    }
    private val transformMode = CoordinateTransformMode.SIMPLE // テスト用
    
    /**
     * プレビューサイズを設定（座標変換用）
     */
    fun setPreviewSize(width: Int, height: Int) {
        previewWidth = width
        previewHeight = height
    }
    
    /**
     * カメラの向きを設定
     */
    fun setCameraFacing(isFrontCamera: Boolean) {
        isUsingFrontCamera = isFrontCamera
    }
    
    /**
     * シンプルな座標変換（サイズ・位置調整付き）
     */
    private fun transformCoordinatesSimple(
        bbox: android.graphics.Rect,
        imageWidth: Int,
        imageHeight: Int
    ): Rect {
        // プレビューサイズが設定されていない場合はデフォルト値を使用
        val displayWidth = if (previewWidth > 0) previewWidth.toFloat() else 1080f
        val displayHeight = if (previewHeight > 0) previewHeight.toFloat() else 1920f
        
        // 基本的なスケール変換
        val scaleX = displayWidth / imageWidth.toFloat()
        val scaleY = displayHeight / imageHeight.toFloat()
        
        // 元のBBOX座標をスケール変換
        val originalLeft = bbox.left * scaleX
        val originalTop = bbox.top * scaleY
        val originalRight = bbox.right * scaleX
        val originalBottom = bbox.bottom * scaleY
        
        // 元のサイズ
        val originalWidth = originalRight - originalLeft
        val originalHeight = originalBottom - originalTop
        
        // 調整後のサイズ
        val adjustedWidth = originalWidth * 1.3f  // 横幅1.3倍
        val adjustedHeight = originalHeight        // 高さはそのまま
        
        // 調整後の位置（中心を基準に横幅を拡大、上に100px移動）
        val centerX = (originalLeft + originalRight) / 2f
        val adjustedLeft = centerX - (adjustedWidth / 2f)
        val adjustedRight = centerX + (adjustedWidth / 2f)
        val adjustedTop = originalTop - 100f  // 100px上に移動
        val adjustedBottom = adjustedTop + adjustedHeight
        
        val result = if (isUsingFrontCamera) {
            // フロントカメラの場合：左右反転
            Rect(
                left = displayWidth - adjustedRight,
                top = adjustedTop,
                right = displayWidth - adjustedLeft,
                bottom = adjustedBottom
            )
        } else {
            // アウトカメラの場合：通常の変換
            Rect(
                left = adjustedLeft,
                top = adjustedTop,
                right = adjustedRight,
                bottom = adjustedBottom
            )
        }
        
        println("BBOX adjustment: scale=${scaleX}x${scaleY}, frontCam=$isUsingFrontCamera")
        println("Original: w=${originalWidth.toInt()}, h=${originalHeight.toInt()}")
        println("Adjusted: w=${adjustedWidth.toInt()}, h=${adjustedHeight.toInt()}, moved up 100px")
        println("Result: $result")
        return result
    }
    
    /**
     * 【保留】複雑な座標変換（後で使用）
     */
    private fun transformCoordinatesComplex(
        bbox: android.graphics.Rect,
        imageWidth: Int,
        imageHeight: Int,
        rotationDegrees: Int
    ): Rect {
        // 複雑な変換ロジック（後で実装）
        return Rect(
            left = bbox.left.toFloat(),
            top = bbox.top.toFloat(),
            right = bbox.right.toFloat(),
            bottom = bbox.bottom.toFloat()
        )
    }

    @androidx.annotation.OptIn(ExperimentalGetImage::class)
    override fun analyze(imageProxy: ImageProxy) {
        val inputImage = InputImage.fromMediaImage(
            imageProxy.image!!,
            imageProxy.imageInfo.rotationDegrees
        )
        
        detector.process(inputImage)
            .addOnSuccessListener { faces ->
                val currentTime = System.currentTimeMillis()
                val results = mutableListOf<FaceDetectionResult>()
                
                // 現在検出されたトラッキングIDのセット
                val currentTrackingIds = mutableSetOf<Int>()
                
                faces.forEach { face ->
                    val trackingId = face.trackingId
                    if (trackingId != null) {
                        currentTrackingIds.add(trackingId)
                        
                        // 座標変換モードに応じて処理を切り替え
                        val boundingBox = when (transformMode) {
                            CoordinateTransformMode.RAW -> {
                                // 生座標（変換なし）
                                Rect(
                                    left = face.boundingBox.left.toFloat(),
                                    top = face.boundingBox.top.toFloat(),
                                    right = face.boundingBox.right.toFloat(),
                                    bottom = face.boundingBox.bottom.toFloat()
                                )
                            }
                            CoordinateTransformMode.SIMPLE -> {
                                // シンプルなスケール変換
                                transformCoordinatesSimple(
                                    face.boundingBox,
                                    imageProxy.width,
                                    imageProxy.height
                                )
                            }
                            CoordinateTransformMode.COMPLEX -> {
                                // 複雑な変換
                                transformCoordinatesComplex(
                                    face.boundingBox,
                                    imageProxy.width,
                                    imageProxy.height,
                                    imageProxy.imageInfo.rotationDegrees
                                )
                            }
                        }
                        
                        // デバッグログ出力
                        println("=== BBOX Debug ===")
                        println("Transform mode: $transformMode")
                        println("Face detected: trackingId=$trackingId")
                        println("Original ML Kit bbox: ${face.boundingBox}")
                        println("Transformed bbox: $boundingBox")
                        println("Image size: ${imageProxy.width}x${imageProxy.height}")
                        println("Preview size: ${previewWidth}x${previewHeight}")
                        println("Rotation: ${imageProxy.imageInfo.rotationDegrees}°")
                        
                        // トラッキング情報を取得または作成
                        val trackingInfo = trackingInfoMap.getOrPut(trackingId) {
                            FaceTrackingInfo(trackingId = trackingId)
                        }
                        
                        // 安定性を更新
                        val centerPoint = Pair(boundingBox.center.x, boundingBox.center.y)
                        val minSize = minOf(boundingBox.width, boundingBox.height)
                        val isStable = trackingInfo.updateStability(centerPoint, minSize, currentTime)
                        
                        // Phase 2: 安定した顔に対して埋め込み抽出と認識処理
                        if (isStable && embeddingExtractor != null && faceRecognizer != null) {
                            processStableFaceForRecognition(face, trackingId, trackingInfo, imageProxy)
                        }
                        
                        // Unknown IDの割り当て（認識されていない場合）
                        if (trackingInfo.recognizedPersonId == null && trackingInfo.unknownId == null && trackingInfo.hasAttemptedRecognition) {
                            trackingInfo.unknownId = unknownIdCounter++
                        }
                        
                        // 認識情報の作成
                        val recognitionInfo = createRecognitionInfo(trackingInfo)
                        
                        val result = FaceDetectionResult(
                            face = face,
                            boundingBox = boundingBox,
                            trackingId = trackingId,
                            isStable = isStable,
                            confidence = 1.0f, // ML Kitは信頼度を直接提供しないため固定値
                            recognitionInfo = recognitionInfo
                        )
                        
                        results.add(result)
                    }
                }
                
                // 検出されなくなったトラッキングIDを削除（タイムアウト処理）
                val expiredTrackingIds = trackingInfoMap.keys.filter { id ->
                    id !in currentTrackingIds && 
                    currentTime - trackingInfoMap[id]!!.lastDetectionTime > 5000L // 5秒でタイムアウト
                }
                expiredTrackingIds.forEach { expiredId ->
                    trackingInfoMap.remove(expiredId)
                    lastEmbeddingExtractionTime.remove(expiredId)
                    lastExtractedEmbeddings.remove(expiredId)
                    clearFaceThumbnail(expiredId)
                }
                
                _detectionResults.value = results
            }
            .addOnFailureListener { exception ->
                // エラーログを出力（実際のアプリではログライブラリを使用）
                println("Face detection failed: ${exception.message}")
                _detectionResults.value = emptyList()
            }
            .addOnCompleteListener {
                imageProxy.close()
            }
    }
    
    /**
     * 指定されたトラッキングIDの顔を既知の人物として認識
     */
    fun recognizeFace(trackingId: Int, personId: Long) {
        trackingInfoMap[trackingId]?.let { trackingInfo ->
            trackingInfo.recognizedPersonId = personId
            trackingInfo.unknownId = null
        }
    }
    
    /**
     * 指定されたトラッキングIDのUnknown IDを取得
     */
    fun getUnknownId(trackingId: Int): Int? {
        return trackingInfoMap[trackingId]?.unknownId
    }
    
    /**
     * 指定されたトラッキングIDの認識済み人物IDを取得
     */
    fun getRecognizedPersonId(trackingId: Int): Long? {
        return trackingInfoMap[trackingId]?.recognizedPersonId
    }
    
    /**
     * トラッキング情報をクリア
     */
    fun clearTrackingInfo() {
        trackingInfoMap.clear()
        unknownIdCounter = 1
        lastEmbeddingExtractionTime.clear()
        lastExtractedEmbeddings.clear()
        clearAllFaceThumbnails()
    }
    
    /**
     * 認識情報を作成
     */
    private fun createRecognitionInfo(trackingInfo: FaceTrackingInfo): FaceRecognitionInfo {
        return if (trackingInfo.recognizedPersonId != null) {
            // キャッシュから人物情報と要約を取得
            val cachedInfo = trackingInfo.cachedPersonInfo
            
            FaceRecognitionInfo(
                personId = trackingInfo.recognizedPersonId,
                personName = cachedInfo?.personName ?: "認識済み",
                unknownId = null,
                confidence = 0.9f, // 仮の信頼度
                lastSummary = cachedInfo?.lastSummary // Phase 4: 要約表示
            )
        } else {
            FaceRecognitionInfo(
                personId = null,
                personName = null,
                unknownId = trackingInfo.unknownId,
                confidence = 0f
            )
        }
    }
    
    /**
     * 顔サムネイルをキャッシュ
     */
    private fun cacheFaceThumbnail(trackingId: Int, bitmap: Bitmap) {
        lastCapturedFaceBitmaps.remove(trackingId)?.recycle()
        lastCapturedFaceBitmaps[trackingId] = bitmap
    }
    
    /**
     * 顔サムネイルのキャッシュを削除
     */
    private fun clearFaceThumbnail(trackingId: Int) {
        lastCapturedFaceBitmaps.remove(trackingId)?.recycle()
    }
    
    /**
     * すべての顔サムネイルキャッシュを削除
     */
    private fun clearAllFaceThumbnails() {
        lastCapturedFaceBitmaps.forEach { (_, bitmap) ->
            bitmap.recycle()
        }
        lastCapturedFaceBitmaps.clear()
    }
    
    /**
     * 顔サムネイルを生成
     */
    private fun captureFaceThumbnail(
        frameBitmap: Bitmap,
        face: com.google.mlkit.vision.face.Face
    ): Bitmap? {
        val bbox = face.boundingBox
        if (bbox.width() <= 0 || bbox.height() <= 0) {
            return null
        }
        
        val widthScale = 1.3f
        val heightScale = 1.35f
        val centerX = bbox.exactCenterX()
        val centerY = bbox.exactCenterY() - bbox.height() * 0.1f
        
        val halfWidth = (bbox.width() * widthScale / 2f)
        val halfHeight = (bbox.height() * heightScale / 2f)
        
        val left = (centerX - halfWidth).roundToInt().coerceAtLeast(0)
        val top = (centerY - halfHeight).roundToInt().coerceAtLeast(0)
        val right = (centerX + halfWidth).roundToInt().coerceAtMost(frameBitmap.width)
        val bottom = (centerY + halfHeight).roundToInt().coerceAtMost(frameBitmap.height)
        
        val finalWidth = right - left
        val finalHeight = bottom - top
        
        if (finalWidth <= 0 || finalHeight <= 0) {
            return null
        }
        
        return try {
            Bitmap.createBitmap(frameBitmap, left, top, finalWidth, finalHeight)
        } catch (e: Exception) {
            println("Failed to crop face thumbnail: ${e.message}")
            null
        }
    }
    
    
    /**
     * 顔サムネイルを保存して人物の代表画像に設定
     */
    private suspend fun persistFaceThumbnailIfAvailable(personId: Long, trackingId: Int) {
        val bitmap = lastCapturedFaceBitmaps.remove(trackingId) ?: return
        try {
            val imagePath = withContext(Dispatchers.IO) {
                FaceImageStorage.saveFaceThumbnail(appContext, personId, bitmap)
            }
            if (imagePath != null) {
                personRepository?.updatePersonProfileImage(personId, imagePath)
            }
        } catch (e: Exception) {
            println("Failed to persist face thumbnail: ${e.message}")
        } finally {
            bitmap.recycle()
        }
    }
    
    /**
     * 既知人物に代表画像が登録されていなければ登録する
     */
    private suspend fun ensureProfileImageForPerson(personId: Long, trackingId: Int) {
        val repository = personRepository ?: run {
            clearFaceThumbnail(trackingId)
            return
        }
        
        try {
            val person = repository.getPersonById(personId)
            if (person?.profileImagePath.isNullOrBlank()) {
                persistFaceThumbnailIfAvailable(personId, trackingId)
            } else {
                clearFaceThumbnail(trackingId)
            }
        } catch (e: Exception) {
            println("Failed to ensure profile image: ${e.message}")
            clearFaceThumbnail(trackingId)
        }
    }
    
    /**
     * 安定した顔に対する認識処理
     */
    private fun processStableFaceForRecognition(
        face: com.google.mlkit.vision.face.Face,
        trackingId: Int,
        trackingInfo: FaceTrackingInfo,
        imageProxy: ImageProxy
    ) {
        val currentTime = System.currentTimeMillis()
        val lastTime = lastEmbeddingExtractionTime[trackingId] ?: 0L
        
        // 埋め込み抽出の間隔制御（1-2秒に1回）
        if (currentTime - lastTime < AppConstants.EMBEDDING_CAPTURE_INTERVAL_MS) {
            return
        }
        
        // 既に認識済みの場合はスキップ（追加の埋め込みは別途実装）
        if (trackingInfo.recognizedPersonId != null) {
            return
        }
        
        lastEmbeddingExtractionTime[trackingId] = currentTime
        
        // 非同期で埋め込み抽出と認識処理を実行
        CoroutineScope(Dispatchers.IO).launch {
            try {
                // ImageProxyからBitmapを作成
                val bitmap = imageProxyToBitmap(imageProxy)
                
                // 顔サムネイルをキャッシュ
                captureFaceThumbnail(bitmap, face)?.let { thumbnail ->
                    cacheFaceThumbnail(trackingId, thumbnail)
                }
                
                // 埋め込み抽出
                val embedding = embeddingExtractor!!.extractEmbedding(bitmap, face)
                if (embedding != null) {
                    // 最後に抽出した埋め込みを保存
                    lastExtractedEmbeddings[trackingId] = embedding
                    
                    // 認識処理
                    val recognitionResult = faceRecognizer!!.recognizeFace(embedding)
                    
                    // 結果をメインスレッドで処理
                    CoroutineScope(Dispatchers.Main).launch {
                        trackingInfo.hasAttemptedRecognition = true
                        when (recognitionResult) {
                            is RecognitionResult.Recognized -> {
                                trackingInfo.recognizedPersonId = recognitionResult.personId
                                trackingInfo.unknownId = null
                                
                                // Phase 4: 人物情報と最新要約をキャッシュ
                                CoroutineScope(Dispatchers.IO).launch {
                                    try {
                                        // 既知人物に新しい埋め込みを追加（品質向上のため）
                                        faceRecognizer.addEmbeddingToExistingPerson(
                                            recognitionResult.personId,
                                            embedding
                                        )
                                        
                                        // 代表画像を保持していなければ保存
                                        ensureProfileImageForPerson(
                                            recognitionResult.personId,
                                            trackingId
                                        )
                                        
                                        // 人物情報と最新要約を取得してキャッシュ
                                        val (person, lastSummary) = personRepository?.getPersonWithLatestSummary(recognitionResult.personId) ?: Pair(null, null)
                                        
                                        trackingInfo.cachedPersonInfo = CachedPersonInfo(
                                            personName = person?.name,
                                            lastSummary = lastSummary
                                        )
                                        
                                    } catch (e: Exception) {
                                        println("Failed to cache person info: ${e.message}")
                                    }
                                }
                            }
                            is RecognitionResult.Unknown -> {
                                // Unknown状態を維持（ユーザーの命名を待つ）
                            }
                        }
                    }
                }
            } catch (e: Exception) {
                println("Failed to process face for recognition: ${e.message}")
            }
        }
    }
    
    /**
     * ImageProxyからBitmapを作成（YUV_420_888 → RGB変換）
     */
    private fun imageProxyToBitmap(imageProxy: ImageProxy): Bitmap {
        val yBuffer = imageProxy.planes[0].buffer
        val uBuffer = imageProxy.planes[1].buffer
        val vBuffer = imageProxy.planes[2].buffer
        
        val ySize = yBuffer.remaining()
        val uSize = uBuffer.remaining()
        val vSize = vBuffer.remaining()
        
        val nv21 = ByteArray(ySize + uSize + vSize)
        
        // Y成分をコピー
        yBuffer.get(nv21, 0, ySize)
        
        // UV成分をインターリーブしてコピー
        val uvPixelStride = imageProxy.planes[1].pixelStride
        if (uvPixelStride == 1) {
            // UV成分が連続している場合
            uBuffer.get(nv21, ySize, uSize)
            vBuffer.get(nv21, ySize + uSize, vSize)
        } else {
            // UV成分がインターリーブされている場合
            val uvBuffer = ByteArray(uSize + vSize)
            uBuffer.get(uvBuffer, 0, uSize)
            vBuffer.get(uvBuffer, uSize, vSize)
            
            var uvIndex = 0
            for (i in 0 until uSize step uvPixelStride) {
                nv21[ySize + uvIndex] = uvBuffer[i]
                nv21[ySize + uvIndex + 1] = uvBuffer[uSize + i]
                uvIndex += 2
            }
        }
        
        // YUV → RGB変換
        return yuvToRgbBitmap(nv21, imageProxy.width, imageProxy.height)
    }
    
    /**
     * YUVデータからRGB Bitmapを作成
     */
    private fun yuvToRgbBitmap(yuv: ByteArray, width: Int, height: Int): Bitmap {
        val pixels = IntArray(width * height)
        val frameSize = width * height
        
        for (j in 0 until height) {
            for (i in 0 until width) {
                val y = yuv[j * width + i].toInt() and 0xFF
                val uvIndex = frameSize + (j / 2) * width + (i and 1.inv())
                val u = yuv[uvIndex].toInt() and 0xFF
                val v = yuv[uvIndex + 1].toInt() and 0xFF
                
                // YUV → RGB変換式
                val r = (y + 1.402 * (v - 128)).toInt().coerceIn(0, 255)
                val g = (y - 0.344 * (u - 128) - 0.714 * (v - 128)).toInt().coerceIn(0, 255)
                val b = (y + 1.772 * (u - 128)).toInt().coerceIn(0, 255)
                
                pixels[j * width + i] = (0xFF shl 24) or (r shl 16) or (g shl 8) or b
            }
        }
        
        return Bitmap.createBitmap(pixels, width, height, Bitmap.Config.ARGB_8888)
    }
    
    /**
     * 新しい人物として保存
     */
    suspend fun saveNewPersonWithEmbedding(trackingId: Int, name: String): Long? {
        val trackingInfo = trackingInfoMap[trackingId] ?: return null
        val embedding = lastExtractedEmbeddings[trackingId] ?: return null
        
        return try {
            val personId = faceRecognizer?.saveNewPerson(name, embedding) ?: return null
            trackingInfo.recognizedPersonId = personId
            trackingInfo.unknownId = null
            trackingInfo.cachedPersonInfo = CachedPersonInfo(
                personName = name,
                lastSummary = null
            )
            
            // 保存済みの埋め込みを削除
            lastExtractedEmbeddings.remove(trackingId)
            lastEmbeddingExtractionTime.remove(trackingId)
            
            // サムネイルを保存
            persistFaceThumbnailIfAvailable(personId, trackingId)
            
            // last_seen_at更新
            personRepository?.updateLastSeenAt(personId)
            
            personId
        } catch (e: Exception) {
            println("Failed to save new person: ${e.message}")
            null
        }
    }
    
    /**
     * リソースの解放
     */
    fun release() {
        detector.close()
        trackingInfoMap.clear()
        lastEmbeddingExtractionTime.clear()
        lastExtractedEmbeddings.clear()
        clearAllFaceThumbnails()
        embeddingExtractor?.release()
    }
}
