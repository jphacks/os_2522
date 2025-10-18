package com.example.daredakke.constants

/**
 * アプリケーション全体で使用する定数を定義
 * 仕様書のしきい値・初期パラメータに対応
 */
object AppConstants {
    
    // カメラ設定
    const val CAMERA_WIDTH = 640
    const val CAMERA_HEIGHT = 480
    
    // 顔検出・安定化
    const val MIN_FACE_SIZE_PX = 120 // BBOX短辺の最小サイズ
    const val MAX_FACE_CENTER_MOVEMENT_PX = 28f // 顔中心の最大移動量
    const val FACE_STABILITY_DURATION_MS = 200L // 顔安定判定の連続時間
    
    // 顔認識
    const val FACE_RECOGNITION_COSINE_THRESHOLD = 0.76f // コサイン類似度しきい値
    const val FACE_RECOGNITION_L2_THRESHOLD = 1.2f // L2距離しきい値
    const val MAX_EMBEDDINGS_PER_PERSON = 10 // 1人あたりの最大埋め込み数
    const val TOP_K_EMBEDDINGS_FOR_MATCHING = 5 // 照合に使用する上位k個の埋め込み
    const val EMBEDDING_CAPTURE_INTERVAL_MS = 1500L // 埋め込み取得間隔（1-2秒）
    
    // 音声・録音
    const val AUDIO_SAMPLE_RATE = 16000 // 16kHz
    const val AUDIO_CHANNEL_COUNT = 1 // mono
    const val AUDIO_BIT_DEPTH = 16 // 16bit PCM
    
    // VAD（Voice Activity Detection）
    const val VAD_SPEECH_START_DURATION_MS = 400L // 発話開始判定の連続時間
    const val VAD_SILENCE_END_DURATION_MS = 2500L // 無音による終了判定時間
    const val FACE_LOSS_END_DURATION_MS = 2000L // 顔喪失による終了判定時間
    
    // 録音バッファ
    const val PREROLL_DURATION_MS = 2500L // プリロール時間（2-3秒）
    const val HANGOVER_DURATION_MS = 3000L // ハングオーバ時間（2-4秒）
    
    // Gemini API
    const val GEMINI_SUMMARY_MAX_CHARS = 25 // 要約の最大文字数
    const val GEMINI_SUMMARY_LANGUAGE = "ja" // 要約言語
    
    // データベース
    const val DATABASE_NAME = "arsome_database"
    const val DATABASE_VERSION = 2
    
    // UI
    const val UNKNOWN_PERSON_PREFIX = "Unknown #"
    const val BBOX_STROKE_WIDTH = 4f
    const val BBOX_TEXT_SIZE = 16f
    
    // ファイル・ディレクトリ
    const val RECORDINGS_DIR = "recordings"
    const val MODELS_DIR = "models"
    const val MOBILEFACENET_MODEL_FILE = "mobilefacenet.tflite"
    
    // ネットワーク・API
    const val API_TIMEOUT_SECONDS = 30L
    const val MAX_RETRY_ATTEMPTS = 3
    const val RETRY_DELAY_MS = 1000L
    
    // セキュリティ
    const val ENCRYPTION_KEY_ALIAS = "ARsomeEncryptionKey"
    const val SHARED_PREFS_NAME = "ARsomePrefs"
    const val SHARED_PREFS_ENCRYPTED_NAME = "ARsomeEncryptedPrefs"
}
