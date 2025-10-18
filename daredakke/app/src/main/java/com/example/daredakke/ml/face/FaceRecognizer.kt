package com.example.daredakke.ml.face

import com.example.daredakke.constants.AppConstants
import com.example.daredakke.data.entities.FaceEmbedding
import com.example.daredakke.data.repository.PersonRepository
import kotlin.math.sqrt

/**
 * 顔認識処理クラス
 * 埋め込みベクトルの照合と既知人物の判定を行う
 */
class FaceRecognizer(
    private val personRepository: PersonRepository
) {
    
    /**
     * 顔埋め込みを既知人物と照合
     */
    suspend fun recognizeFace(embedding: FloatArray): RecognitionResult {
        val allEmbeddings = personRepository.getAllEmbeddings()

        println("📊 Recognition: Total stored embeddings=${allEmbeddings.size}, Persons=${allEmbeddings.groupBy { it.personId }.size}")

        if (allEmbeddings.isEmpty()) {
            println("⚠️ No stored embeddings found")
            return RecognitionResult.Unknown
        }

        // 人物ごとにグループ化
        val embeddingsByPerson = allEmbeddings.groupBy { it.personId }

        var bestMatch: Pair<Long, Float>? = null

        embeddingsByPerson.forEach { (personId, personEmbeddings) ->
            // 上位k個の埋め込みを使用（仕様書の安定化処理）
            val topEmbeddings = personEmbeddings
                .sortedByDescending { it.createdAt }
                .take(AppConstants.TOP_K_EMBEDDINGS_FOR_MATCHING)

            // 各埋め込みとの類似度を計算
            val similarities = topEmbeddings.map { storedEmbedding ->
                val storedVector = byteArrayToFloatArray(storedEmbedding.vector)
                calculateCosineSimilarity(embedding, storedVector)
            }

            // 上位k個の平均類似度
            val avgSimilarity = similarities.average().toFloat()

            println("👤 PersonId=$personId: avgSimilarity=$avgSimilarity (from ${similarities.size} embeddings)")

            // 最高スコアの更新
            if (bestMatch == null || avgSimilarity > bestMatch!!.second) {
                bestMatch = Pair(personId, avgSimilarity)
            }
        }

        return bestMatch?.let { (personId, similarity) ->
            if (similarity > AppConstants.FACE_RECOGNITION_COSINE_THRESHOLD) {
                println("✅ Match found: PersonId=$personId, similarity=$similarity (threshold=${AppConstants.FACE_RECOGNITION_COSINE_THRESHOLD})")
                RecognitionResult.Recognized(personId, similarity)
            } else {
                println("❌ Best match below threshold: PersonId=$personId, similarity=$similarity (threshold=${AppConstants.FACE_RECOGNITION_COSINE_THRESHOLD})")
                RecognitionResult.Unknown
            }
        } ?: RecognitionResult.Unknown
    }
    
    /**
     * 新しい人物として埋め込みを保存
     */
    suspend fun saveNewPerson(name: String, embedding: FloatArray): Long {
        // 新しい人物を作成
        val personId = personRepository.createPersonWithName(name)
        
        // 埋め込みを保存
        val faceEmbedding = FaceEmbedding(
            personId = personId,
            vector = floatArrayToByteArray(embedding)
        )
        personRepository.insertEmbedding(faceEmbedding)
        
        return personId
    }
    
    /**
     * 既知人物に新しい埋め込みを追加
     */
    suspend fun addEmbeddingToExistingPerson(personId: Long, embedding: FloatArray) {
        val faceEmbedding = FaceEmbedding(
            personId = personId,
            vector = floatArrayToByteArray(embedding)
        )
        personRepository.insertEmbedding(faceEmbedding)
        
        // last_seen_atを更新
        personRepository.updateLastSeenAt(personId)
    }
    
    /**
     * コサイン類似度の計算
     */
    private fun calculateCosineSimilarity(a: FloatArray, b: FloatArray): Float {
        if (a.size != b.size) return 0f
        
        var dotProduct = 0f
        var normA = 0f
        var normB = 0f
        
        for (i in a.indices) {
            dotProduct += a[i] * b[i]
            normA += a[i] * a[i]
            normB += b[i] * b[i]
        }
        
        val denominator = sqrt(normA) * sqrt(normB)
        return if (denominator > 0f) dotProduct / denominator else 0f
    }
    
    /**
     * L2距離の計算（オプション）
     */
    private fun calculateL2Distance(a: FloatArray, b: FloatArray): Float {
        if (a.size != b.size) return Float.MAX_VALUE
        
        var sum = 0f
        for (i in a.indices) {
            val diff = a[i] - b[i]
            sum += diff * diff
        }
        
        return sqrt(sum)
    }
    
    /**
     * FloatArray -> ByteArray 変換
     */
    private fun floatArrayToByteArray(floatArray: FloatArray): ByteArray {
        val byteArray = ByteArray(floatArray.size * 4)
        var index = 0
        
        for (float in floatArray) {
            val bits = java.lang.Float.floatToIntBits(float)
            byteArray[index++] = (bits and 0xFF).toByte()
            byteArray[index++] = ((bits shr 8) and 0xFF).toByte()
            byteArray[index++] = ((bits shr 16) and 0xFF).toByte()
            byteArray[index++] = ((bits shr 24) and 0xFF).toByte()
        }
        
        return byteArray
    }
    
    /**
     * ByteArray -> FloatArray 変換
     */
    private fun byteArrayToFloatArray(byteArray: ByteArray): FloatArray {
        val floatArray = FloatArray(byteArray.size / 4)
        var byteIndex = 0
        
        for (i in floatArray.indices) {
            val bits = (byteArray[byteIndex].toInt() and 0xFF) or
                      ((byteArray[byteIndex + 1].toInt() and 0xFF) shl 8) or
                      ((byteArray[byteIndex + 2].toInt() and 0xFF) shl 16) or
                      ((byteArray[byteIndex + 3].toInt() and 0xFF) shl 24)
            
            floatArray[i] = java.lang.Float.intBitsToFloat(bits)
            byteIndex += 4
        }
        
        return floatArray
    }
}

/**
 * 顔認識結果を表すsealed class
 */
sealed class RecognitionResult {
    object Unknown : RecognitionResult()
    data class Recognized(val personId: Long, val confidence: Float) : RecognitionResult()
}
