package com.example.daredakke.ml.audio

import android.media.AudioFormat
import android.media.AudioRecord
import android.media.MediaRecorder
import com.example.daredakke.constants.AppConstants
import kotlinx.coroutines.*
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import java.io.File
import java.io.FileOutputStream
import java.io.IOException
import java.nio.ByteBuffer
import java.nio.ByteOrder
import java.util.concurrent.ConcurrentLinkedQueue

/**
 * 音声録音クラス
 * 仕様書のPhase 3に対応：プリロール/ハングオーバ機能付き録音
 */
class AudioRecorder {
    
    companion object {
        private const val BUFFER_SIZE_MULTIPLIER = 4
    }
    
    // 録音設定（仕様書準拠：16kHz/mono/16bit）
    private val sampleRate = AppConstants.AUDIO_SAMPLE_RATE
    private val channelConfig = AudioFormat.CHANNEL_IN_MONO
    private val audioFormat = AudioFormat.ENCODING_PCM_16BIT
    private val bufferSize = AudioRecord.getMinBufferSize(sampleRate, channelConfig, audioFormat) * BUFFER_SIZE_MULTIPLIER
    
    private var audioRecord: AudioRecord? = null
    private var recordingJob: Job? = null
    
    // 録音状態
    private val _isRecording = MutableStateFlow(false)
    val isRecording: StateFlow<Boolean> = _isRecording.asStateFlow()
    
    private val _recordingFile = MutableStateFlow<File?>(null)
    val recordingFile: StateFlow<File?> = _recordingFile.asStateFlow()
    
    // プリロールバッファ（リングバッファ）
    private val prerollBuffer = ConcurrentLinkedQueue<ShortArray>()
    private val prerollMaxSize = (AppConstants.PREROLL_DURATION_MS * sampleRate / 1000 / 1024).toInt() // 約2-3秒分
    
    // 録音データ
    private var currentRecordingData = mutableListOf<ShortArray>()
    private var recordingStartTime: Long = 0
    
    /**
     * 録音の初期化
     */
    fun initialize(): Boolean {
        return try {
            if (audioRecord?.state == AudioRecord.STATE_INITIALIZED) {
                return true
            }
            
            audioRecord = AudioRecord(
                MediaRecorder.AudioSource.MIC,
                sampleRate,
                channelConfig,
                audioFormat,
                bufferSize
            )
            
            audioRecord?.state == AudioRecord.STATE_INITIALIZED
        } catch (e: SecurityException) {
            println("AudioRecord initialization failed: Permission denied")
            false
        } catch (e: Exception) {
            println("AudioRecord initialization failed: ${e.message}")
            false
        }
    }
    
    /**
     * 連続音声監視の開始（プリロールバッファ用）
     */
    fun startContinuousListening(
        onAudioFrame: (ShortArray) -> Unit
    ) {
        if (audioRecord?.state != AudioRecord.STATE_INITIALIZED) {
            if (!initialize()) {
                println("Failed to initialize AudioRecord for continuous listening")
                return
            }
        }
        
        recordingJob?.cancel()
        recordingJob = CoroutineScope(Dispatchers.IO).launch {
            val buffer = ShortArray(1024) // 約64msのバッファ
            
            try {
                audioRecord?.startRecording()
                
                while (isActive) {
                    val bytesRead = audioRecord?.read(buffer, 0, buffer.size) ?: 0
                    
                    if (bytesRead > 0) {
                        val audioFrame = buffer.copyOf(bytesRead)
                        
                        // プリロールバッファの管理
                        updatePrerollBuffer(audioFrame)
                        
                        // 音声フレームコールバック
                        onAudioFrame(audioFrame)
                    }
                    
                    delay(10) // CPU負荷軽減
                }
            } catch (e: Exception) {
                println("Continuous listening error: ${e.message}")
            } finally {
                audioRecord?.stop()
            }
        }
    }
    
    /**
     * プリロールバッファの更新
     */
    private fun updatePrerollBuffer(audioFrame: ShortArray) {
        prerollBuffer.offer(audioFrame)
        
        // バッファサイズ制限
        while (prerollBuffer.size > prerollMaxSize) {
            prerollBuffer.poll()
        }
    }
    
    /**
     * 録音開始（プリロールデータを含む）
     */
    fun startRecording(outputFile: File) {
        if (_isRecording.value) {
            println("Recording already in progress")
            return
        }
        
        recordingStartTime = System.currentTimeMillis()
        synchronized(currentRecordingData) {
            currentRecordingData.clear()

            // プリロールデータを追加
            currentRecordingData.addAll(prerollBuffer.toList())
        }

        _isRecording.value = true
        _recordingFile.value = outputFile
        
        println("Recording started with preroll: ${prerollBuffer.size} frames")
    }
    
    /**
     * 録音中の音声フレーム追加
     */
    fun addRecordingFrame(audioFrame: ShortArray) {
        if (_isRecording.value) {
            synchronized(currentRecordingData) {
                currentRecordingData.add(audioFrame)
            }
        }
    }
    
    /**
     * 録音停止（ハングオーバ期間後）
     */
    fun stopRecording() {
        if (!_isRecording.value) {
            println("No recording in progress")
            return
        }
        
        CoroutineScope(Dispatchers.IO).launch {
            // ハングオーバ期間の録音継続
            delay(AppConstants.HANGOVER_DURATION_MS)

            // 録音状態を先に終了して、新しいフレームの追加を停止
            _isRecording.value = false

            // スナップショットを作成してConcurrentModificationExceptionを回避
            val recordingSnapshot = synchronized(currentRecordingData) {
                currentRecordingData.toList()
            }
            val frameCount = recordingSnapshot.size

            val outputFile = _recordingFile.value
            if (outputFile != null && recordingSnapshot.isNotEmpty()) {
                saveRecordingToWav(outputFile, recordingSnapshot)
            }

            _recordingFile.value = null
            synchronized(currentRecordingData) {
                currentRecordingData.clear()
            }

            val duration = System.currentTimeMillis() - recordingStartTime
            println("Recording stopped. Duration: ${duration}ms, Frames: ${frameCount}")
        }
    }
    
    /**
     * WAVファイルとして保存
     */
    private fun saveRecordingToWav(outputFile: File, audioData: List<ShortArray>) {
        try {
            val totalSamples = audioData.sumOf { it.size }
            val dataSize = totalSamples * 2 // 16bit = 2 bytes per sample
            val fileSize = dataSize + 44 // WAVヘッダーサイズ
            
            FileOutputStream(outputFile).use { fos ->
                // WAVヘッダーを書き込み
                writeWavHeader(fos, totalSamples, sampleRate, dataSize, fileSize)
                
                // 音声データを書き込み
                val byteBuffer = ByteBuffer.allocate(dataSize).order(ByteOrder.LITTLE_ENDIAN)
                audioData.forEach { frame ->
                    frame.forEach { sample ->
                        byteBuffer.putShort(sample)
                    }
                }
                
                fos.write(byteBuffer.array())
            }
            
            println("Recording saved to: ${outputFile.absolutePath}")
        } catch (e: IOException) {
            println("Failed to save recording: ${e.message}")
        }
    }
    
    /**
     * WAVヘッダーの書き込み
     */
    private fun writeWavHeader(
        fos: FileOutputStream,
        totalSamples: Int,
        sampleRate: Int,
        dataSize: Int,
        fileSize: Int
    ) {
        val header = ByteBuffer.allocate(44).order(ByteOrder.LITTLE_ENDIAN)
        
        // RIFF header
        header.put("RIFF".toByteArray())
        header.putInt(fileSize - 8)
        header.put("WAVE".toByteArray())
        
        // fmt chunk
        header.put("fmt ".toByteArray())
        header.putInt(16) // chunk size
        header.putShort(1) // audio format (PCM)
        header.putShort(1) // num channels (mono)
        header.putInt(sampleRate)
        header.putInt(sampleRate * 2) // byte rate
        header.putShort(2) // block align
        header.putShort(16) // bits per sample
        
        // data chunk
        header.put("data".toByteArray())
        header.putInt(dataSize)
        
        fos.write(header.array())
    }
    
    /**
     * 連続監視の停止
     */
    fun stopContinuousListening() {
        recordingJob?.cancel()
        audioRecord?.stop()
        prerollBuffer.clear()
    }
    
    /**
     * リソースの解放
     */
    fun release() {
        stopContinuousListening()
        audioRecord?.release()
        audioRecord = null
        synchronized(currentRecordingData) {
            currentRecordingData.clear()
        }
        prerollBuffer.clear()
    }
    
    /**
     * デバッグ情報
     */
    fun getDebugInfo(): AudioRecorderDebugInfo {
        val frameCount = synchronized(currentRecordingData) {
            currentRecordingData.size
        }
        return AudioRecorderDebugInfo(
            isRecording = _isRecording.value,
            prerollBufferSize = prerollBuffer.size,
            currentRecordingFrames = frameCount,
            recordingDuration = if (_isRecording.value) {
                System.currentTimeMillis() - recordingStartTime
            } else 0L
        )
    }
}

/**
 * AudioRecorderデバッグ情報
 */
data class AudioRecorderDebugInfo(
    val isRecording: Boolean,
    val prerollBufferSize: Int,
    val currentRecordingFrames: Int,
    val recordingDuration: Long
)
