package com.example.daredakke.ml.face

import android.content.Context
import android.graphics.Bitmap
import android.graphics.Canvas
import android.graphics.Matrix
import android.graphics.Paint
import android.graphics.PointF
import com.google.mlkit.vision.face.Face
import com.google.mlkit.vision.face.FaceLandmark
import org.tensorflow.lite.Interpreter
import org.tensorflow.lite.support.common.FileUtil
import java.nio.ByteBuffer
import java.nio.ByteOrder
import kotlin.math.cos
import kotlin.math.sin
import kotlin.math.sqrt

/**
 * MobileFaceNetを使用した顔埋め込み抽出クラス
 * 仕様書のPhase 2に対応：128D埋め込み生成と5点ランドマークアライメント
 */
class FaceEmbeddingExtractor(private val context: Context) {
    
    companion object {
        private const val MODEL_FILE_NAME = "models/mobilefacenet.tflite"
        private const val INPUT_SIZE = 112 // 112x112 input
        private const val EMBEDDING_SIZE = 192 // 128D output
        private const val PIXEL_SIZE = 3 // RGB
        private const val BATCH_SIZE = 2
    }
    
    private var interpreter: Interpreter? = null
    private var isInitialized = false
    
    /**
     * モデルの初期化
     */
    suspend fun initialize(): Boolean {
        return try {
            if (isInitialized) return true

            // アセットからモデルファイルを読み込み
            val modelBuffer = FileUtil.loadMappedFile(context, MODEL_FILE_NAME)

            // Interpreter作成（GPU使用も検討可能）
            val options = Interpreter.Options().apply {
                setNumThreads(4) // CPU使用時のスレッド数
                setUseNNAPI(true) // Android Neural Networks API使用
            }

            interpreter = Interpreter(modelBuffer, options)

            // モデルの入力形状を確認してログ出力
            val inputTensor = interpreter!!.getInputTensor(0)
            val inputShape = inputTensor.shape()
            val inputDataType = inputTensor.dataType()
            println("Model input shape: ${inputShape.contentToString()}")
            println("Model input data type: $inputDataType")
            println("Expected input size: ${inputShape[1]} x ${inputShape[2]}")

            isInitialized = true
            true
        } catch (e: Exception) {
            println("Failed to initialize MobileFaceNet: ${e.message}")
            false
        }
    }
    
    /**
     * 顔画像から128D埋め込みを抽出
     */
    fun extractEmbedding(faceBitmap: Bitmap, face: Face): FloatArray? {
        if (!isInitialized || interpreter == null) {
            println("FaceEmbeddingExtractor not initialized")
            return null
        }
        
        return try {
            // 5点ランドマークによるアライメント
            val alignedBitmap = alignFace(faceBitmap, face) ?: return null
            
            // 112x112にリサイズ
            val resizedBitmap = Bitmap.createScaledBitmap(
                alignedBitmap, 
                INPUT_SIZE, 
                INPUT_SIZE, 
                true
            )
            
            // 前処理：正規化とバッファ変換
            val inputBuffer = preprocessBitmap(resizedBitmap)
            
            // 推論実行
            val outputArray = Array(BATCH_SIZE) { FloatArray(EMBEDDING_SIZE) }
            interpreter!!.run(inputBuffer, outputArray)
            
            // L2正規化
            val embedding = outputArray[0]
            normalizeL2(embedding)
            
            embedding
        } catch (e: Exception) {
            println("Failed to extract face embedding: ${e.message}")
            null
        }
    }
    
    /**
     * 5点ランドマークによる顔のアライメント
     */
    private fun alignFace(bitmap: Bitmap, face: Face): Bitmap? {
        val landmarks = face.allLandmarks
        
        // 必要なランドマークを取得
        val leftEye = landmarks.find { it.landmarkType == FaceLandmark.LEFT_EYE }?.position
        val rightEye = landmarks.find { it.landmarkType == FaceLandmark.RIGHT_EYE }?.position
        val nose = landmarks.find { it.landmarkType == FaceLandmark.NOSE_BASE }?.position
        val leftMouth = landmarks.find { it.landmarkType == FaceLandmark.MOUTH_LEFT }?.position
        val rightMouth = landmarks.find { it.landmarkType == FaceLandmark.MOUTH_RIGHT }?.position
        
        // ランドマークが不十分な場合はバウンディングボックスのみ使用
        if (leftEye == null || rightEye == null) {
            val bbox = face.boundingBox
            return Bitmap.createBitmap(
                bitmap,
                maxOf(0, bbox.left),
                maxOf(0, bbox.top),
                minOf(bitmap.width - maxOf(0, bbox.left), bbox.width()),
                minOf(bitmap.height - maxOf(0, bbox.top), bbox.height())
            )
        }
        
        // 目の中心を基準とした回転角度計算
        val eyeAngle = kotlin.math.atan2(
            (rightEye.y - leftEye.y).toDouble(),
            (rightEye.x - leftEye.x).toDouble()
        )
        
        // アフィン変換行列の作成
        val matrix = Matrix()
        val centerX = (leftEye.x + rightEye.x) / 2f
        val centerY = (leftEye.y + rightEye.y) / 2f
        
        // 回転補正
        matrix.setRotate(Math.toDegrees(-eyeAngle).toFloat(), centerX, centerY)
        
        // アライメント済みビットマップを作成
        return try {
            Bitmap.createBitmap(bitmap, 0, 0, bitmap.width, bitmap.height, matrix, true)
        } catch (e: Exception) {
            println("Failed to align face: ${e.message}")
            bitmap // アライメント失敗時は元の画像を返す
        }
    }
    
    /**
     * ビットマップを前処理してByteBufferに変換
     */
    private fun preprocessBitmap(bitmap: Bitmap): ByteBuffer {
        val inputBuffer = ByteBuffer.allocateDirect(
            BATCH_SIZE * INPUT_SIZE * INPUT_SIZE * PIXEL_SIZE * 4
        ).apply {
            order(ByteOrder.nativeOrder())
        }
        
        val pixels = IntArray(INPUT_SIZE * INPUT_SIZE)
        bitmap.getPixels(pixels, 0, INPUT_SIZE, 0, 0, INPUT_SIZE, INPUT_SIZE)
        
        for (pixel in pixels) {
            // RGB値を抽出して正規化 [0, 255] -> [-1, 1]
            val r = ((pixel shr 16) and 0xFF) / 127.5f - 1.0f
            val g = ((pixel shr 8) and 0xFF) / 127.5f - 1.0f
            val b = (pixel and 0xFF) / 127.5f - 1.0f
            
            inputBuffer.putFloat(r)
            inputBuffer.putFloat(g)
            inputBuffer.putFloat(b)
        }
        
        return inputBuffer
    }
    
    /**
     * L2正規化
     */
    private fun normalizeL2(embedding: FloatArray) {
        val norm = sqrt(embedding.map { it * it }.sum())
        if (norm > 0) {
            for (i in embedding.indices) {
                embedding[i] /= norm
            }
        }
    }
    
    /**
     * リソースの解放
     */
    fun release() {
        interpreter?.close()
        interpreter = null
        isInitialized = false
    }
}
