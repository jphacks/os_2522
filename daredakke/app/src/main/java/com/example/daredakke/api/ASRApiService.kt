package com.example.daredakke.api

import retrofit2.Response
import retrofit2.http.Body
import retrofit2.http.POST
import retrofit2.http.Header

/**
 * Google Speech-to-Text APIサービス
 * 仕様書のPhase 4に対応：音声ファイルの文字起こし
 */
interface ASRApiService {
    
    /**
     * Google Speech-to-Text APIで音声を文字起こし
     * @param request 音声認識リクエスト
     * @return 文字起こし結果
     */
    @POST("./speech:recognize")
    suspend fun recognizeSpeech(
        @Header("Authorization") authorization: String,
        @Body request: GoogleSpeechRequest
    ): Response<GoogleSpeechResponse>
}

/**
 * Google Speech-to-Text リクエスト
 */
data class GoogleSpeechRequest(
    val config: RecognitionConfig,
    val audio: RecognitionAudio
)

data class RecognitionConfig(
    val encoding: String = "LINEAR16",
    val sampleRateHertz: Int = 16000,
    val languageCode: String = "ja-JP",
    val maxAlternatives: Int = 1,
    val enableAutomaticPunctuation: Boolean = true
)

data class RecognitionAudio(
    val content: String // Base64エンコードされた音声データ
)

/**
 * Google Speech-to-Text レスポンス
 */
data class GoogleSpeechResponse(
    val results: List<SpeechRecognitionResult>?
)

data class SpeechRecognitionResult(
    val alternatives: List<SpeechRecognitionAlternative>
)

data class SpeechRecognitionAlternative(
    val transcript: String,
    val confidence: Float
)

/**
 * 統一されたASRレスポンス（既存コードとの互換性のため）
 */
data class ASRResponse(
    val text: String,
    val confidence: Float? = null,
    val language: String? = null,
    val duration: Float? = null,
    val error: String? = null
)

/**
 * ASR API エラーレスポンス
 */
data class ASRErrorResponse(
    val error: String,
    val code: Int,
    val message: String
)
