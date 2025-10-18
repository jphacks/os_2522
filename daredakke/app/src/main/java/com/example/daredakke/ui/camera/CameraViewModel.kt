package com.example.daredakke.ui.camera

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.daredakke.daredakkeApplication
import com.example.daredakke.ml.face.FaceDetectionResult
import com.example.daredakke.ml.face.FaceDetector
import com.example.daredakke.ml.face.FaceEmbeddingExtractor
import com.example.daredakke.ml.face.FaceRecognizer
import com.example.daredakke.ml.audio.VADProcessor
import com.example.daredakke.ml.audio.AudioRecorder
import com.example.daredakke.ml.audio.ThreeLayerGating
import com.example.daredakke.ml.audio.RecordingTrigger
import com.example.daredakke.workers.TranscriptionWorker
import androidx.work.OneTimeWorkRequestBuilder
import androidx.work.WorkManager
import androidx.work.workDataOf
import java.io.File
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * カメラ画面のViewModel
 * Phase 1: 顔検出結果の管理とUI状態の制御
 * Phase 2: 顔認識機能とデータベース統合
 */
class CameraViewModel(application: Application) : AndroidViewModel(application) {
    
    private val app = application as daredakkeApplication
    private val personRepository = app.personRepository
    
    // Phase 2: ML コンポーネント
    private var faceEmbeddingExtractor: FaceEmbeddingExtractor? = null
    private var faceRecognizer: FaceRecognizer? = null
    private var faceDetector: FaceDetector? = null
    
    // Phase 3: 音声・録音コンポーネント
    private var vadProcessor: VADProcessor? = null
    private var audioRecorder: AudioRecorder? = null
    private var threeLayerGating: ThreeLayerGating? = null
    
    // Phase 3: 録音ファイル管理
    private val recordingsDir by lazy {
        File(getApplication<Application>().filesDir, "recordings").apply {
            if (!exists()) mkdirs()
        }
    }
    
    // Phase 4: 録音セッション管理
    private var currentRecordingSession: RecordingSession? = null
    
    init {
        // Phase 2&3コンポーネントの初期化
        initializePhase2Components()
        initializePhase3Components()
    }
    
    private val _detectionResults = MutableStateFlow<List<FaceDetectionResult>>(emptyList())
    val detectionResults: StateFlow<List<FaceDetectionResult>> = _detectionResults.asStateFlow()
    
    private val _isRecording = MutableStateFlow(false)
    val isRecording: StateFlow<Boolean> = _isRecording.asStateFlow()
    
    private val _showNameDialog = MutableStateFlow(false)
    val showNameDialog: StateFlow<Boolean> = _showNameDialog.asStateFlow()
    
    private val _isUsingFrontCamera = MutableStateFlow(false) // デフォルトはアウトカメラ
    val isUsingFrontCamera: StateFlow<Boolean> = _isUsingFrontCamera.asStateFlow()
    
    /**
     * Phase 2コンポーネントの初期化
     */
    private fun initializePhase2Components() {
        viewModelScope.launch {
            try {
                // 埋め込み抽出器の初期化
                faceEmbeddingExtractor = FaceEmbeddingExtractor(getApplication()).apply {
                    if (!initialize()) {
                        println("Warning: FaceEmbeddingExtractor initialization failed")
                    }
                }
                
                // 顔認識器の初期化
                faceRecognizer = FaceRecognizer(personRepository)
                
            } catch (e: Exception) {
                println("Failed to initialize Phase 2 components: ${e.message}")
            }
        }
    }
    
    /**
     * Phase 3コンポーネントの初期化
     */
    private fun initializePhase3Components() {
        viewModelScope.launch {
            try {
                // VADプロセッサの初期化
                vadProcessor = VADProcessor()
                
                // オーディオレコーダーの初期化
                audioRecorder = AudioRecorder().apply {
                    if (!initialize()) {
                        println("Warning: AudioRecorder initialization failed")
                    }
                }
                
                // 三層ゲーティングシステムの初期化
                threeLayerGating = ThreeLayerGating()
                
                // 三層ゲーティングの監視開始
                startThreeLayerGatingMonitoring()
                
                // 連続音声監視の開始
                startContinuousAudioMonitoring()
                
            } catch (e: Exception) {
                println("Failed to initialize Phase 3 components: ${e.message}")
            }
        }
    }
    
    /**
     * 顔検出器を設定し、検出結果の監視を開始
     */
    fun setFaceDetector(detector: FaceDetector) {
        faceDetector = detector
        
        viewModelScope.launch {
            detector.detectionResults.collect { results ->
                _detectionResults.value = results
                
                // Phase 3: 視覚トリガーの更新
                val hasStableFace = results.any { it.isStable }
                threeLayerGating?.updateVisualTrigger(hasStableFace)
            }
        }
    }
    
    /**
     * Phase 2&4統合用の顔検出器を作成
     */
    fun createIntegratedFaceDetector(): FaceDetector {
        return FaceDetector(
            appContext = getApplication(),
            embeddingExtractor = faceEmbeddingExtractor,
            faceRecognizer = faceRecognizer,
            personRepository = personRepository // Phase 4: 要約表示のため
        )
    }
    
    /**
     * 登録ボタンがタップされた時の処理
     */
    fun onRegisterButtonTapped() {
        val stableFace = getStableUnknownFace()
        if (stableFace != null) {
            _showNameDialog.value = true
        }
    }

    /**
     * 名前入力ダイアログを閉じる
     */
    fun dismissNameDialog() {
        _showNameDialog.value = false
    }

    /**
     * 未認識の顔を取得（安定性条件を緩和）
     */
    private fun getStableUnknownFace() = _detectionResults.value.firstOrNull {
        it.recognitionInfo?.isRecognized != true
    }

    /**
     * 新しい人物として名前を保存（Phase 2完全実装）
     */
    fun savePersonName(name: String) {
        viewModelScope.launch {
            try {
                // 最も安定した未認識の顔を取得
                val stableFace = getStableUnknownFace()
                val trackingId = stableFace?.trackingId

                if (trackingId != null) {
                    println("Attempting to save person with trackingId: $trackingId, name: $name")
                    // 顔検出器に新しい人物として保存を依頼
                    val personId = faceDetector?.saveNewPersonWithEmbedding(trackingId, name)

                    if (personId != null) {
                        println("Successfully saved new person: $name (ID: $personId)")
                    } else {
                        println("Failed to save new person: $name. personId is null.")
                    }
                } else {
                    println("No stable unknown face found to save")
                }

                dismissNameDialog()
            } catch (e: Exception) {
                println("Failed to save person name: ${e.message}")
                dismissNameDialog()
            }
        }
    }
    
    /**
     * 三層ゲーティング監視の開始
     */
    private fun startThreeLayerGatingMonitoring() {
        viewModelScope.launch {
            threeLayerGating?.recordingTrigger?.collect { trigger ->
                when (trigger) {
                    RecordingTrigger.START -> {
                        startRecording()
                    }
                    RecordingTrigger.STOP -> {
                        stopRecording()
                    }
                    null -> {
                        // トリガーなし
                    }
                }
            }
        }
    }
    
    /**
     * 連続音声監視の開始
     */
    private fun startContinuousAudioMonitoring() {
        audioRecorder?.startContinuousListening { audioFrame ->
            // VAD処理
            val voiceActivity = vadProcessor?.processAudioFrame(audioFrame, 16000)
            
            if (voiceActivity != null) {
                // 三層ゲーティングに音声トリガーを通知
                threeLayerGating?.updateAudioTrigger(voiceActivity)
            }
            
            // 録音中の場合は音声フレームを追加
            if (_isRecording.value) {
                audioRecorder?.addRecordingFrame(audioFrame)
            }
        }
    }
    
    /**
     * 録音開始
     */
    private fun startRecording() {
        if (_isRecording.value) {
            return
        }
        
        val timestamp = System.currentTimeMillis()
        val filename = "recording_$timestamp.wav"
        val outputFile = File(recordingsDir, filename)
        
        // Phase 4: 録音セッション情報を記録
        val recognizedPersonId = _detectionResults.value
            .find { it.isStable }
            ?.recognitionInfo
            ?.personId
        
        currentRecordingSession = RecordingSession(
            startTime = timestamp,
            outputFile = outputFile,
            personId = recognizedPersonId
        )
        
        audioRecorder?.startRecording(outputFile)
        _isRecording.value = true
        
        println("Recording started: $filename, personId: $recognizedPersonId")
    }
    
    /**
     * 録音停止
     */
    private fun stopRecording() {
        if (!_isRecording.value) {
            return
        }
        
        val session = currentRecordingSession
        audioRecorder?.stopRecording()
        _isRecording.value = false
        
        // Phase 4: 文字起こし・要約ジョブを開始
        if (session != null) {
            scheduleTranscriptionJob(session)
        }
        
        currentRecordingSession = null
        println("Recording stopped and transcription job scheduled")
    }
    
    /**
     * Phase 4: 文字起こし・要約ジョブのスケジュール
     */
    private fun scheduleTranscriptionJob(session: RecordingSession) {
        val workRequest = OneTimeWorkRequestBuilder<TranscriptionWorker>()
            .setInputData(
                workDataOf(
                    TranscriptionWorker.KEY_AUDIO_FILE_PATH to session.outputFile.absolutePath,
                    TranscriptionWorker.KEY_PERSON_ID to (session.personId ?: -1L),
                    TranscriptionWorker.KEY_ENCOUNTER_ID to -1L,
                    TranscriptionWorker.KEY_STARTED_AT to session.startTime,
                    TranscriptionWorker.KEY_ENDED_AT to System.currentTimeMillis()
                )
            )
            .build()
        
        WorkManager.getInstance(getApplication()).enqueue(workRequest)
        
        println("Transcription job scheduled for personId: ${session.personId}")
    }
    
    /**
     * 録音状態の更新（手動制御用）
     */
    fun updateRecordingState(isRecording: Boolean) {
        if (isRecording) {
            startRecording()
        } else {
            stopRecording()
        }
    }
    
    /**
     * カメラの向きを切り替え
     */
    fun toggleCamera() {
        _isUsingFrontCamera.value = !_isUsingFrontCamera.value
    }
    
    override fun onCleared() {
        super.onCleared()
        faceDetector?.release()
        faceEmbeddingExtractor?.release()
        audioRecorder?.release()
        threeLayerGating?.reset()
    }
}

/**
 * 録音セッション情報
 */
private data class RecordingSession(
    val startTime: Long,
    val outputFile: File,
    val personId: Long?
)
