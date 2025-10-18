package com.example.daredakke.api

import com.example.daredakke.api.ASRApiService
import com.example.daredakke.BuildConfig
import com.example.daredakke.constants.AppConstants
import okhttp3.Interceptor
import okhttp3.OkHttpClient
import okhttp3.logging.HttpLoggingInterceptor
import retrofit2.Retrofit
import retrofit2.converter.gson.GsonConverterFactory
import java.util.concurrent.TimeUnit

/**
 * API クライアント
 * ASR・Gemini API への接続を管理
 */
object ApiClient {

    // Google Speech-to-Text API
    private const val ASR_BASE_URL = "https://speech.googleapis.com/v1/"

    // Gemini API (正式なエンドポイント)
    private const val GEMINI_BASE_URL = "https://generativelanguage.googleapis.com/v1beta/"

    private val loggingInterceptor = HttpLoggingInterceptor().apply {
        level = HttpLoggingInterceptor.Level.BODY
    }

    // Gemini API用の認証インターセプター
    private val geminiAuthInterceptor = Interceptor { chain ->
        val originalRequest = chain.request()
        val authenticatedRequest = originalRequest.newBuilder()
            .addHeader("x-goog-api-key", BuildConfig.GEMINI_API_KEY)
            .addHeader("Content-Type", "application/json")
            .build()
        chain.proceed(authenticatedRequest)
    }

    // ASR API用の基本OkHttpClient
    private val asrOkHttpClient = OkHttpClient.Builder()
        .addInterceptor(loggingInterceptor)
        .connectTimeout(AppConstants.API_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .readTimeout(AppConstants.API_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .writeTimeout(AppConstants.API_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    // Gemini API用の認証付きOkHttpClient
    private val geminiOkHttpClient = OkHttpClient.Builder()
        .addInterceptor(loggingInterceptor)
        .addInterceptor(geminiAuthInterceptor)
        .connectTimeout(AppConstants.API_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .readTimeout(AppConstants.API_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .writeTimeout(AppConstants.API_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    private val asrRetrofit = Retrofit.Builder()
        .baseUrl(ASR_BASE_URL)
        .client(asrOkHttpClient)
        .addConverterFactory(GsonConverterFactory.create())
        .build()

    private val geminiRetrofit = Retrofit.Builder()
        .baseUrl(GEMINI_BASE_URL)
        .client(geminiOkHttpClient)
        .addConverterFactory(GsonConverterFactory.create())
        .build()

    val asrApiService: ASRApiService = asrRetrofit.create(ASRApiService::class.java)
    val geminiApiService: GeminiApiService = geminiRetrofit.create(GeminiApiService::class.java)
}

/**
 * API結果を表すsealed class
 */
sealed class ApiResult<T> {
    data class Success<T>(val data: T) : ApiResult<T>()
    data class Error<T>(val exception: Exception) : ApiResult<T>()
    data class NetworkError<T>(val message: String) : ApiResult<T>()
}

/**
 * API呼び出しのヘルパー関数
 */
suspend fun <T> safeApiCall(apiCall: suspend () -> retrofit2.Response<T>): ApiResult<T> {
    return try {
        val response = apiCall()
        if (response.isSuccessful) {
            response.body()?.let { body ->
                ApiResult.Success(body)
            } ?: ApiResult.Error(Exception("Response body is null"))
        } else {
            ApiResult.NetworkError("HTTP ${response.code()}: ${response.message()}")
        }
    } catch (e: Exception) {
        ApiResult.Error(e)
    }
}
