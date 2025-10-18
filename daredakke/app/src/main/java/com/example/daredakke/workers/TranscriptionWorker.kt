package com.example.daredakke.workers

import android.content.Context
import androidx.work.CoroutineWorker
import androidx.work.WorkerParameters
import androidx.work.workDataOf
import com.example.daredakke.daredakkeApplication
import com.example.daredakke.data.entities.Encounter
import com.example.daredakke.service.TranscriptionResult
import com.example.daredakke.service.TranscriptionService
import java.io.File

/**
 * 文字起こし・要約処理のワーカー
 * 仕様書のPhase 4に対応：非同期ジョブによるASR・Gemini処理
 */
class TranscriptionWorker(
    context: Context,
    params: WorkerParameters
) : CoroutineWorker(context, params) {
    
    companion object {
        const val KEY_AUDIO_FILE_PATH = "audio_file_path"
        const val KEY_PERSON_ID = "person_id"
        const val KEY_ENCOUNTER_ID = "encounter_id"
        const val KEY_STARTED_AT = "started_at"
        const val KEY_ENDED_AT = "ended_at"
        // 結果キー
        const val KEY_RESULT_SUCCESS = "result_success"
        const val KEY_RESULT_TRANSCRIPTION = "result_transcription"
        const val KEY_RESULT_SUMMARY = "result_summary"
        const val KEY_RESULT_ERROR = "result_error"
    }
    
    private val transcriptionService = TranscriptionService(applicationContext)
    private val personRepository = (applicationContext as daredakkeApplication).personRepository
    
    override suspend fun doWork(): Result {
        val audioFilePath = inputData.getString(KEY_AUDIO_FILE_PATH)
            ?: return Result.failure(workDataOf(KEY_RESULT_ERROR to "Audio file path not provided"))
        
        val personId = inputData.getLong(KEY_PERSON_ID, -1L).takeIf { it != -1L }
        val encounterId = inputData.getLong(KEY_ENCOUNTER_ID, -1L)
        val startedAt = inputData.getLong(KEY_STARTED_AT, 0L)
        val endedAt = inputData.getLong(KEY_ENDED_AT, 0L)
        
        val audioFile = File(audioFilePath)
        if (!audioFile.exists()) {
            return Result.failure(workDataOf(KEY_RESULT_ERROR to "Audio file not found"))
        }
        
        return try {
            when (val transcriptionResult = transcriptionService.processAudioToSummary(audioFile)) {
                is TranscriptionResult.Success -> {
                    saveEncounterToDatabase(
                        personId = personId,
                        encounterId = encounterId,
                        startedAt = startedAt,
                        endedAt = endedAt,
                        transcription = transcriptionResult.transcription,
                        summary = transcriptionResult.summary
                    )
                    personId?.let { personRepository.updateLastSeenAt(it) }
                    Result.success(
                        workDataOf(
                            KEY_RESULT_SUCCESS to true,
                            KEY_RESULT_TRANSCRIPTION to transcriptionResult.transcription,
                            KEY_RESULT_SUMMARY to transcriptionResult.summary
                        )
                    )
                }
                is TranscriptionResult.PartialSuccess -> {
                    saveEncounterToDatabase(
                        personId = personId,
                        encounterId = encounterId,
                        startedAt = startedAt,
                        endedAt = endedAt,
                        transcription = transcriptionResult.transcription,
                        summary = null
                    )
                    personId?.let { personRepository.updateLastSeenAt(it) }
                    Result.success(
                        workDataOf(
                            KEY_RESULT_SUCCESS to false,
                            KEY_RESULT_TRANSCRIPTION to transcriptionResult.transcription,
                            KEY_RESULT_ERROR to transcriptionResult.error
                        )
                    )
                }
                is TranscriptionResult.EmptyTranscription -> {
                    Result.success(
                        workDataOf(
                            KEY_RESULT_SUCCESS to false,
                            KEY_RESULT_ERROR to "音声が検出されませんでした"
                        )
                    )
                }
                is TranscriptionResult.TranscriptionError -> {
                    Result.failure(
                        workDataOf(
                            KEY_RESULT_ERROR to transcriptionResult.error
                        )
                    )
                }
            }
        } catch (e: Exception) {
            Result.failure(
                workDataOf(
                    KEY_RESULT_ERROR to (e.message ?: "不明なエラーが発生しました")
                )
            )
        } finally {
            if (audioFile.exists()) {
                audioFile.delete()
            }
        }
    }
    
    /**
     * エンカウンターをデータベースに保存
     */
    private suspend fun saveEncounterToDatabase(
        personId: Long?,
        encounterId: Long,
        startedAt: Long,
        endedAt: Long,
        transcription: String,
        summary: String?
    ) {
        try {
            if (encounterId != -1L) {
                // 既存のエンカウンターを更新
                val existingEncounter = personRepository.getEncounterById(encounterId)
                if (existingEncounter != null) {
                    val updatedEncounter = existingEncounter.copy(
                        asrText = transcription,
                        summaryText = summary
                    )
                    personRepository.updateEncounter(updatedEncounter)
                }
            } else {
                // 新しいエンカウンターを作成
                val encounter = Encounter(
                    personId = personId,
                    startedAt = startedAt,
                    endedAt = endedAt,
                    asrText = transcription,
                    summaryText = summary
                )
                personRepository.insertEncounter(encounter)
            }
            
            println("Encounter saved successfully: personId=$personId, summary=${summary?.take(20)}...")
        } catch (e: Exception) {
            println("Failed to save encounter: ${e.message}")
            throw e
        }
    }
}
