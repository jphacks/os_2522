package com.example.daredakke.api

import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.POST

/**
 * Gemini API サービス
 * 実際のGemini API (generateContent) エンドポイントに対応
 */
interface GeminiApiService {
    
    /**
     * Gemini APIでテキスト生成（要約）
     * @param modelName モデル名 (例: "gemini-1.5-flash")
     * @param request 生成リクエスト
     * @return 生成結果
     */
    @POST("models/{model}:generateContent")
    suspend fun generateContent(
        @retrofit2.http.Path("model") modelName: String = "gemini-1.5-flash",
        @Body request: GeminiGenerateRequest
    ): Response<GeminiGenerateResponse>
}

/**
 * Gemini生成リクエスト
 */
data class GeminiGenerateRequest(
    val contents: List<GeminiContent>,
    val generationConfig: GeminiGenerationConfig? = null
)

data class GeminiContent(
    val parts: List<GeminiPart>
)

data class GeminiPart(
    val text: String
)

data class GeminiGenerationConfig(
    val temperature: Float = 0.7f,
    val topK: Int = 40,
    val topP: Float = 0.95f,
    val maxOutputTokens: Int = 100,
    val candidateCount: Int = 1
)

/**
 * Gemini生成レスポンス
 */
data class GeminiGenerateResponse(
    val candidates: List<GeminiCandidate>?,
    val error: GeminiError? = null
)

data class GeminiCandidate(
    val content: GeminiContent?,
    val finishReason: String?,
    val safetyRatings: List<GeminiSafetyRating>?
)

data class GeminiSafetyRating(
    val category: String,
    val probability: String
)

/**
 * 要約用のヘルパー関数で使用するレスポンス
 */
data class GeminiSummarizeResponse(
    val summary: String,
    val originalLength: Int? = null,
    val summaryLength: Int? = null,
    val confidence: Float? = null,
    val introducedName: String? = null,
    val nameConfidence: Float? = null,
    val error: String? = null
)

/**
 * Gemini API エラー
 */
data class GeminiError(
    val code: Int,
    val message: String,
    val status: String
)

/**
 * Gemini API エラーレスポンス
 */
data class GeminiErrorResponse(
    val error: GeminiError
)
