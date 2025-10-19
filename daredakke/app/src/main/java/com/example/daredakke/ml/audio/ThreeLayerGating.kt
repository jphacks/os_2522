package com.example.daredakke.ml.audio

import com.example.daredakke.constants.AppConstants
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.delay
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * 三層ゲーティングシステム
 * 仕様書のPhase 3に対応：視覚×音声×安定化による自動録音制御
 */
class ThreeLayerGating {
    
    // ゲーティング状態
    private val _gatingState = MutableStateFlow(GatingState.INACTIVE)
    val gatingState: StateFlow<GatingState> = _gatingState.asStateFlow()
    
    private val _recordingTrigger = MutableStateFlow<RecordingTrigger?>(null)
    val recordingTrigger: StateFlow<RecordingTrigger?> = _recordingTrigger.asStateFlow()
    
    // 各層の状態
    private var visualTrigger = false
    private var audioTrigger = VoiceActivityState.SILENCE
    private var stabilizationState = StabilizationState.UNSTABLE
    
    // タイムスタンプ管理
    private var lastVisualTriggerTime: Long = 0
    private var lastAudioTriggerTime: Long = 0
    private var recordingStartTime: Long? = null
    
    /**
     * 第1層：視覚トリガーの更新
     * @param hasFaceStable 安定した顔が検出されているか
     */
    fun updateVisualTrigger(hasFaceStable: Boolean) {
        val currentTime = System.currentTimeMillis()
        
        if (hasFaceStable) {
            if (!visualTrigger) {
                visualTrigger = true
                lastVisualTriggerTime = currentTime
                println("Visual trigger ACTIVATED: Stable face detected")
            }
        } else {
            if (visualTrigger) {
                visualTrigger = false
                println("Visual trigger DEACTIVATED: No stable face")
            }
        }
        
        updateGatingState()
    }
    
    /**
     * 第2層：音声トリガーの更新
     * @param voiceActivity VADプロセッサからの音声活動状態
     */
    fun updateAudioTrigger(voiceActivity: VoiceActivityState) {
        val currentTime = System.currentTimeMillis()
        val previousState = audioTrigger
        
        audioTrigger = voiceActivity
        
        if (voiceActivity == VoiceActivityState.SPEECH && previousState != VoiceActivityState.SPEECH) {
            lastAudioTriggerTime = currentTime
            println("Audio trigger ACTIVATED: Speech detected")
        }
        
        updateGatingState()
    }
    
    /**
     * 第3層：安定化状態の更新
     */
    private fun updateStabilizationState() {
        val currentTime = System.currentTimeMillis()
        
        stabilizationState = when {
            // 両方のトリガーが有効で、視覚トリガーが最近有効になった
            visualTrigger && audioTrigger == VoiceActivityState.SPEECH -> {
                val visualAge = currentTime - lastVisualTriggerTime
                val audioAge = currentTime - lastAudioTriggerTime
                
                if (visualAge <= 5000 && audioAge <= 200) { // 視覚5秒、音声2秒以内
                    println("stabilizationState  visualAge <= 5000 && audioAge <= 200")
                    StabilizationState.STABLE
                } else {
                    println("stabilizationState  NOT visualAge <= 5000 && audioAge <= 200")
                    StabilizationState.STABLE
                }
            }
            
            // 片方のみ有効
            visualTrigger || audioTrigger == VoiceActivityState.SPEECH -> {
                StabilizationState.PARTIAL
                //StabilizationState.STABLE
            }
            
            // 両方とも無効
            else -> StabilizationState.UNSTABLE
        }
    }
    
    /**
     * ゲーティング状態の更新と録音制御判定
     */
    private fun updateGatingState() {
        updateStabilizationState()
        
        val currentTime = System.currentTimeMillis()
        val previousState = _gatingState.value
        
        val newState = when (stabilizationState) {
            StabilizationState.STABLE -> {
                // 録音開始条件：安定状態で録音中でない場合
                if (previousState == GatingState.INACTIVE) {
                    recordingStartTime = currentTime
                    _recordingTrigger.value = RecordingTrigger.START
                    GatingState.RECORDING
                } else {
                    GatingState.RECORDING
                }
            }
            
            StabilizationState.PARTIAL -> {
                // 録音中なら継続、そうでなければ待機
                if (previousState == GatingState.RECORDING) {
                    GatingState.RECORDING
                } else {
                    GatingState.INACTIVE
                    //GatingState.WAITING
                }
            }
            
            StabilizationState.UNSTABLE -> {
                // 録音停止条件の判定
                if (previousState == GatingState.RECORDING) {
                    val shouldStop = checkRecordingStopConditions(currentTime)
                    if (shouldStop) {
                        _recordingTrigger.value = RecordingTrigger.STOP
                        recordingStartTime = null
                        GatingState.INACTIVE
                    } else {
                        GatingState.RECORDING
                    }
                } else {
                    GatingState.INACTIVE
                }
            }
        }
        
        // 状態変更の通知
        if (newState != previousState) {
            _gatingState.value = newState
            println("Gating state changed: $previousState -> $newState")
            
            // トリガーのクリア（一度だけ発火）
            if (_recordingTrigger.value != null) {
                CoroutineScope(Dispatchers.Main).launch {
                    delay(100) // UIが反応する時間を確保
                    _recordingTrigger.value = null
                }
            }
        }
    }
    
    /**
     * 録音停止条件の判定
     */
    private fun checkRecordingStopConditions(currentTime: Long): Boolean {
        val recordingDuration = recordingStartTime?.let { currentTime - it } ?: 0L
        
        // 最小録音時間（1秒）以下の場合は停止しない
        if (recordingDuration < 1000) {
            return false
        }
        
        // 顔喪失による停止判定
        val faceAge = currentTime - lastVisualTriggerTime
        val faceBasedStop = !visualTrigger && faceAge >= AppConstants.FACE_LOSS_END_DURATION_MS
        
        // 音声無音による停止判定
        val audioBasedStop = audioTrigger == VoiceActivityState.SPEECH_END
        
        // 最大録音時間（30秒）による強制停止
        val maxDurationStop = recordingDuration >= 30000
        
        val shouldStop = faceBasedStop || audioBasedStop || maxDurationStop
        
        if (shouldStop) {
            val reason = when {
                maxDurationStop -> "maximum duration"
                faceBasedStop -> "face loss timeout"
                audioBasedStop -> "speech end detected"
                else -> "unknown"
            }
            println("Recording stop triggered by: $reason")
        }
        
        return shouldStop
    }
    
    /**
     * デバッグ情報の取得
     */
    fun getDebugInfo(): ThreeLayerGatingDebugInfo {
        val currentTime = System.currentTimeMillis()
        
        return ThreeLayerGatingDebugInfo(
            gatingState = _gatingState.value,
            visualTrigger = visualTrigger,
            audioTrigger = audioTrigger,
            stabilizationState = stabilizationState,
            visualAge = currentTime - lastVisualTriggerTime,
            audioAge = currentTime - lastAudioTriggerTime,
            recordingDuration = recordingStartTime?.let { currentTime - it }
        )
    }
    
    /**
     * システムのリセット
     */
    fun reset() {
        visualTrigger = false
        audioTrigger = VoiceActivityState.SILENCE
        stabilizationState = StabilizationState.UNSTABLE
        _gatingState.value = GatingState.INACTIVE
        _recordingTrigger.value = null
        recordingStartTime = null
        lastVisualTriggerTime = 0
        lastAudioTriggerTime = 0
    }
}

/**
 * ゲーティング状態
 */
enum class GatingState {
    INACTIVE,   // 非活性（録音なし）
    WAITING,    // 待機中（条件部分満足）
    RECORDING   // 録音中
}

/**
 * 録音トリガー
 */
enum class RecordingTrigger {
    START,  // 録音開始
    STOP    // 録音停止
}

/**
 * 安定化状態
 */
enum class StabilizationState {
    UNSTABLE,   // 不安定（条件未満足）
    PARTIAL,    // 部分的（片方の条件のみ満足）
    STABLE      // 安定（両方の条件満足）
}

/**
 * 三層ゲーティングデバッグ情報
 */
data class ThreeLayerGatingDebugInfo(
    val gatingState: GatingState,
    val visualTrigger: Boolean,
    val audioTrigger: VoiceActivityState,
    val stabilizationState: StabilizationState,
    val visualAge: Long,
    val audioAge: Long,
    val recordingDuration: Long?
)
