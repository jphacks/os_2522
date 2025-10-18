package com.example.daredakke.service

import android.content.Context
import android.util.Base64
import com.example.daredakke.api.ApiClient
import com.example.daredakke.api.ApiResult
import com.example.daredakke.api.ASRResponse
import com.example.daredakke.api.GeminiContent
import com.example.daredakke.api.GeminiGenerateRequest
import com.example.daredakke.api.GeminiGenerateResponse
import com.example.daredakke.api.GeminiGenerationConfig
import com.example.daredakke.api.GeminiPart
import com.example.daredakke.api.GeminiSummarizeResponse
import com.example.daredakke.api.GoogleSpeechRequest
import com.example.daredakke.api.RecognitionAudio
import com.example.daredakke.api.RecognitionConfig
import com.example.daredakke.api.safeApiCall
import com.example.daredakke.constants.AppConstants
import com.example.daredakke.BuildConfig
import com.google.auth.oauth2.GoogleCredentials
import org.json.JSONObject
import kotlinx.coroutines.delay
import java.io.ByteArrayOutputStream
import java.io.File
import java.io.InputStream

/**
 * 文字起こし・要約サービス
 * 仕様書のPhase 4に対応：ASR→Gemini要約のパイプライン
 */
class TranscriptionService(private val context: Context) {
    
    private val asrApi = ApiClient.asrApiService
    private val geminiApi = ApiClient.geminiApiService
    
    /**
     * 音声ファイルから要約まで一括処理
     * @param audioFile 録音済みWAVファイル
     * @return 処理結果（文字起こし + 要約）
     */
    suspend fun processAudioToSummary(audioFile: File): TranscriptionResult {
        return try {
            // ステップ1: ASR（文字起こし）
            val asrResult = transcribeAudio(audioFile)
            when (asrResult) {
                is ApiResult.Success -> {
                    val transcription = asrResult.data.text
                    
                    if (transcription.isBlank()) {
                        return TranscriptionResult.EmptyTranscription
                    }
                    
                    // ステップ2: Gemini要約
                    val summaryResult = summarizeText(transcription)
                    when (summaryResult) {
                        is ApiResult.Success -> {
                            TranscriptionResult.Success(
                                transcription = transcription,
                                summary = summaryResult.data.summary,
                                confidence = asrResult.data.confidence ?: 0f,
                                introducedName = summaryResult.data.introducedName,
                                nameDetectionConfidence = summaryResult.data.nameConfidence
                            )
                        }
                        is ApiResult.Error -> {
                            // 要約失敗でも文字起こしは成功
                            TranscriptionResult.PartialSuccess(
                                transcription = transcription,
                                error = summaryResult.exception.message ?: "要約処理失敗"
                            )
                        }
                        is ApiResult.NetworkError -> {
                            TranscriptionResult.PartialSuccess(
                                transcription = transcription,
                                error = summaryResult.message
                            )
                        }
                    }
                }
                is ApiResult.Error -> {
                    TranscriptionResult.TranscriptionError(asrResult.exception.message ?: "文字起こし失敗")
                }
                is ApiResult.NetworkError -> {
                    TranscriptionResult.TranscriptionError(asrResult.message)
                }
            }
        } catch (e: Exception) {
            TranscriptionResult.TranscriptionError(e.message ?: "処理中にエラーが発生しました")
        }
    }
    
    /**
     * 音声ファイルを文字起こし（リトライ付き）
     */
    private suspend fun transcribeAudio(audioFile: File): ApiResult<ASRResponse> {
        var lastException: Exception? = null
        
        repeat(AppConstants.MAX_RETRY_ATTEMPTS) { attempt ->
            try {
                // 1) 認証トークン取得（Service Account JSON）
                val accessToken = getGoogleAccessTokenFromAssets()
                
                // 2) 音声ファイルをBase64化
                val audioBytes = audioFile.readBytes()
                val audioBase64 = Base64.encodeToString(audioBytes, Base64.NO_WRAP)
                
                // 3) リクエスト作成
                val request = GoogleSpeechRequest(
                    config = RecognitionConfig(
                        encoding = "LINEAR16",
                        sampleRateHertz = AppConstants.AUDIO_SAMPLE_RATE,
                        languageCode = "ja-JP",
                        maxAlternatives = 1,
                        enableAutomaticPunctuation = true
                    ),
                    audio = RecognitionAudio(
                        content = audioBase64
                    )
                )
                
                // 4) 呼び出し
                val result: ApiResult<com.example.daredakke.api.GoogleSpeechResponse> = safeApiCall {
                    asrApi.recognizeSpeech(
                        authorization = "Bearer $accessToken",
                        request = request
                    )
                }
                
                if (result is ApiResult.Success) {
                    val response = result.data
                    val transcript = response.results
                        ?.firstOrNull()
                        ?.alternatives
                        ?.firstOrNull()
                        ?.transcript
                        ?: ""
                    val confidence = response.results
                        ?.firstOrNull()
                        ?.alternatives
                        ?.firstOrNull()
                        ?.confidence
                    return ApiResult.Success(ASRResponse(text = transcript, confidence = confidence))
                }
                
                // リトライ前の待機
                if (attempt < AppConstants.MAX_RETRY_ATTEMPTS - 1) {
                    delay(AppConstants.RETRY_DELAY_MS * (attempt + 1))
                }
                
                lastException = when (result) {
                    is ApiResult.Error<*> -> result.exception
                    is ApiResult.NetworkError<*> -> Exception(result.message)
                    else -> Exception("Unknown error")
                }
            } catch (e: Exception) {
                lastException = e
                if (attempt < AppConstants.MAX_RETRY_ATTEMPTS - 1) {
                    delay(AppConstants.RETRY_DELAY_MS * (attempt + 1))
                }
            }
        }
        
        return ApiResult.Error(lastException ?: Exception("ASR failed after retries"))
    }

    private fun getGoogleAccessTokenFromAssets(): String {
        val fileName = BuildConfig.GOOGLE_SPEECH_JSON_KEY_FILE
        require(fileName.isNotBlank()) { "GOOGLE_SPEECH_JSON_KEY_FILE is not set" }
        val inputStream: InputStream = context.assets.open(fileName)
        val credentials = GoogleCredentials.fromStream(inputStream)
            .createScoped(listOf("https://www.googleapis.com/auth/cloud-platform"))
        credentials.refreshIfExpired()
        return credentials.accessToken.tokenValue
    }
    
    /**
     * テキストを要約（リトライ付き）
     */
    private suspend fun summarizeText(text: String): ApiResult<GeminiSummarizeResponse> {
        var lastException: Exception? = null
        
        repeat(AppConstants.MAX_RETRY_ATTEMPTS) { attempt ->
            try {
                // 要約用のプロンプトを作成
                val summaryPrompt = """
                    あなたは会話内容を整理するアシスタントです。以下の会話ログを読み、次の形式のJSONだけを返してください:
                    {
                      "summary": "会話の要点を日本語で${AppConstants.GEMINI_SUMMARY_MAX_CHARS}文字以内にまとめた文",
                      "introducedName": "会話で自分の名前を名乗った人物の名前。誰も名乗っていなければ null",
                      "nameConfidence": 0.0～1.0 の小数（introducedName をどれくらい確信しているか）
                    }
                    - 挨拶や雑談は summary に含めない
                    - introducedName は名乗り全体（例: "田中太郎"）を返す。複数人が名乗った場合は最も本人の名乗りとして確からしいものを1人分だけ返す
                    - JSON 以外のテキストや解説、コードブロックを付けない
                    
                    会話内容:
                    $text
                """.trimIndent()
                
                val request = GeminiGenerateRequest(
                    contents = listOf(
                        GeminiContent(
                            parts = listOf(GeminiPart(summaryPrompt))
                        )
                    ),
                    generationConfig = GeminiGenerationConfig(
                        temperature = 0.3f, // 要約なので低めの温度
                        maxOutputTokens = 50, // 短い要約なので制限
                        topP = 0.8f,
                        topK = 10
                    )
                )
                
                val result = safeApiCall { geminiApi.generateContent("gemini-1.5-flash", request) }
                
                if (result is ApiResult.Success) {
                    val structured = parseStructuredSummaryFromGeminiResponse(result.data)
                    if (structured != null && structured.summary.isNotBlank()) {
                        return ApiResult.Success(
                            structured.copy(
                                originalLength = text.length,
                                summaryLength = structured.summary.length,
                                confidence = structured.confidence ?: 0.9f
                            )
                        )
                    } else {
                        lastException = Exception("Failed to parse structured summary from Gemini response")
                    }
                }
                
                // リトライ前の待機
                if (attempt < AppConstants.MAX_RETRY_ATTEMPTS - 1) {
                    delay(AppConstants.RETRY_DELAY_MS * (attempt + 1)) // 指数バックオフ
                }
                
                lastException = when (result) {
                    is ApiResult.Error<*> -> result.exception
                    is ApiResult.NetworkError<*> -> Exception(result.message)
                    else -> Exception("Unknown error")
                }
                
            } catch (e: Exception) {
                lastException = e
                if (attempt < AppConstants.MAX_RETRY_ATTEMPTS - 1) {
                    delay(AppConstants.RETRY_DELAY_MS * (attempt + 1))
                }
            }
        }
        
        return ApiResult.Error(lastException ?: Exception("Summarization failed after retries"))
    }
    
    /**
     * Gemini APIレスポンスから要約・名乗り情報を抽出
     */
    private fun parseStructuredSummaryFromGeminiResponse(response: GeminiGenerateResponse): GeminiSummarizeResponse? {
        return try {
            val rawText = response.candidates
                ?.firstOrNull()
                ?.content
                ?.parts
                ?.firstOrNull()
                ?.text
                ?.trim()
                ?: return null
            
            val jsonText = rawText
                .removePrefix("```json")
                .removePrefix("```JSON")
                .removeSuffix("```")
                .trim()
            
            val json = JSONObject(jsonText)
            
            val summary = json.optString("summary", "")
            if (summary.isBlank()) {
                return null
            }
            
            val introducedName = when {
                json.isNull("introducedName") -> null
                else -> json.optString("introducedName").takeIf { it.isNotBlank() }
            }
            
            val nameConfidence = if (json.has("nameConfidence")) {
                json.optDouble("nameConfidence", Double.NaN).takeIf { !it.isNaN() }?.toFloat()
            } else {
                null
            }
            
            GeminiSummarizeResponse(
                summary = summary,
                introducedName = introducedName,
                nameConfidence = nameConfidence,
                confidence = null
            )
        } catch (e: Exception) {
            println("Failed to parse structured summary: ${e.message}")
            null
        }
    }
}

/**
 * 文字起こし・要約の結果
 */
sealed class TranscriptionResult {
    data class Success(
        val transcription: String,
        val summary: String,
        val confidence: Float,
        val introducedName: String?,
        val nameDetectionConfidence: Float?
    ) : TranscriptionResult()
    
    data class PartialSuccess(
        val transcription: String,
        val error: String,
        val introducedName: String? = null,
        val nameDetectionConfidence: Float? = null
    ) : TranscriptionResult()
    
    object EmptyTranscription : TranscriptionResult()
    
    data class TranscriptionError(
        val error: String
    ) : TranscriptionResult()
}
