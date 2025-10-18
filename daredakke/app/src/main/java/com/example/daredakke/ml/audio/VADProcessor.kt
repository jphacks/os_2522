package com.example.daredakke.ml.audio

import com.example.daredakke.constants.AppConstants
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlin.math.abs
import kotlin.math.sqrt

/**
 * Voice Activity Detection (VAD) プロセッサ
 * 仕様書のPhase 3に対応：音声トリガーによる録音制御
 */
class VADProcessor {
    
    private val _voiceActivityState = MutableStateFlow(VoiceActivityState.SILENCE)
    val voiceActivityState: StateFlow<VoiceActivityState> = _voiceActivityState.asStateFlow()
    
    // VAD状態管理
    private var speechStartTime: Long? = null
    private var silenceStartTime: Long? = null
    private var currentState = VoiceActivityState.SILENCE
    
    // エネルギーベースVADのパラメータ
    private var noiseFloor = 0.0
    private var adaptiveThreshold = 0.0
    private val energyHistory = mutableListOf<Double>()
    private val maxHistorySize = 50
    
    // 初期化フラグ
    private var isInitialized = false
    private var initFrameCount = 0
    private val initFrameTarget = 30 // 初期30フレームでノイズフロア学習
    
    /**
     * 音声フレームを処理してVAD判定を行う
     * @param audioData 16bit PCM音声データ
     * @param sampleRate サンプルレート
     * @return 現在の音声活動状態
     */
    fun processAudioFrame(audioData: ShortArray, sampleRate: Int): VoiceActivityState {
        val currentTime = System.currentTimeMillis()
        
        // エネルギー計算
        val energy = calculateFrameEnergy(audioData)
        
        // 初期化フェーズ：ノイズフロアの学習
        if (!isInitialized) {
            initializeNoiseFloor(energy)
            return currentState
        }
        
        // 適応的しきい値の更新
        updateAdaptiveThreshold(energy)
        
        // VAD判定
        val isSpeech = energy > adaptiveThreshold
        
        // 状態遷移ロジック
        return updateVoiceActivityState(isSpeech, currentTime)
    }
    
    /**
     * フレームエネルギーを計算
     */
    private fun calculateFrameEnergy(audioData: ShortArray): Double {
        var energy = 0.0
        for (sample in audioData) {
            energy += (sample.toDouble() * sample.toDouble())
        }
        return energy / audioData.size
    }
    
    /**
     * 初期ノイズフロアの学習
     */
    private fun initializeNoiseFloor(energy: Double) {
        energyHistory.add(energy)
        initFrameCount++
        
        if (initFrameCount >= initFrameTarget) {
            // 初期フレームの平均をノイズフロアとする
            noiseFloor = energyHistory.average()
            adaptiveThreshold = noiseFloor * 3.0 // 初期しきい値
            isInitialized = true
            println("VAD initialized: noise floor = $noiseFloor, threshold = $adaptiveThreshold")
        }
    }
    
    /**
     * 適応的しきい値の更新
     */
    private fun updateAdaptiveThreshold(energy: Double) {
        energyHistory.add(energy)
        if (energyHistory.size > maxHistorySize) {
            energyHistory.removeAt(0)
        }
        
        // 現在の状態に応じてしきい値を調整
        when (currentState) {
            VoiceActivityState.SILENCE -> {
                // 無音時は低エネルギーフレームでノイズフロアを更新
                if (energy < adaptiveThreshold * 0.5) {
                    noiseFloor = noiseFloor * 0.95 + energy * 0.05
                    adaptiveThreshold = noiseFloor * 3.0
                }
            }
            VoiceActivityState.SPEECH -> {
                // 発話中はしきい値を若干上げて安定化
                adaptiveThreshold = noiseFloor * 3.5
            }
            VoiceActivityState.SPEECH_END -> {
                // 発話終了検出時はしきい値を下げて継続検出しやすく
                adaptiveThreshold = noiseFloor * 2.5
            }
        }
    }
    
    /**
     * 音声活動状態の更新
     */
    private fun updateVoiceActivityState(isSpeech: Boolean, currentTime: Long): VoiceActivityState {
        val newState = when (currentState) {
            VoiceActivityState.SILENCE -> {
                if (isSpeech) {
                    if (speechStartTime == null) {
                        speechStartTime = currentTime
                    }
                    
                    // 発話開始判定（連続400ms以上）
                    val speechDuration = currentTime - (speechStartTime ?: currentTime)
                    if (speechDuration >= AppConstants.VAD_SPEECH_START_DURATION_MS) {
                        silenceStartTime = null
                        VoiceActivityState.SPEECH
                    } else {
                        VoiceActivityState.SILENCE
                    }
                } else {
                    speechStartTime = null
                    VoiceActivityState.SILENCE
                }
            }
            
            VoiceActivityState.SPEECH -> {
                if (!isSpeech) {
                    if (silenceStartTime == null) {
                        silenceStartTime = currentTime
                    }
                    
                    // 無音継続時間チェック（2500ms）
                    val silenceDuration = currentTime - (silenceStartTime ?: currentTime)
                    if (silenceDuration >= AppConstants.VAD_SILENCE_END_DURATION_MS) {
                        speechStartTime = null
                        VoiceActivityState.SPEECH_END
                    } else {
                        VoiceActivityState.SPEECH
                    }
                } else {
                    silenceStartTime = null
                    VoiceActivityState.SPEECH
                }
            }
            
            VoiceActivityState.SPEECH_END -> {
                // 一度発話終了になったら、次の発話検出まで無音状態
                speechStartTime = null
                silenceStartTime = null
                VoiceActivityState.SILENCE
            }
        }
        
        // 状態変更時の処理
        if (newState != currentState) {
            currentState = newState
            _voiceActivityState.value = newState
            println("VAD state changed to: $newState")
        }
        
        return newState
    }
    
    /**
     * VADプロセッサのリセット
     */
    fun reset() {
        currentState = VoiceActivityState.SILENCE
        speechStartTime = null
        silenceStartTime = null
        _voiceActivityState.value = VoiceActivityState.SILENCE
        
        // 学習データは保持（ノイズフロアの継続学習のため）
    }
    
    /**
     * デバッグ情報の取得
     */
    fun getDebugInfo(): VADDebugInfo {
        return VADDebugInfo(
            currentState = currentState,
            noiseFloor = noiseFloor,
            adaptiveThreshold = adaptiveThreshold,
            isInitialized = isInitialized,
            speechStartTime = speechStartTime,
            silenceStartTime = silenceStartTime
        )
    }
}

/**
 * 音声活動状態
 */
enum class VoiceActivityState {
    SILENCE,    // 無音
    SPEECH,     // 発話中
    SPEECH_END  // 発話終了検出
}

/**
 * VADデバッグ情報
 */
data class VADDebugInfo(
    val currentState: VoiceActivityState,
    val noiseFloor: Double,
    val adaptiveThreshold: Double,
    val isInitialized: Boolean,
    val speechStartTime: Long?,
    val silenceStartTime: Long?
)
