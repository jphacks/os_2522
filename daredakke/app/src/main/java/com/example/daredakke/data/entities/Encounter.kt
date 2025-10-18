package com.example.daredakke.data.entities

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Index
import androidx.room.PrimaryKey

/**
 * 会話エンカウンターエンティティ
 * 仕様書のDBスキーマ: TABLE encounters に対応
 */
@Entity(
    tableName = "encounters",
    foreignKeys = [
        ForeignKey(
            entity = Person::class,
            parentColumns = ["id"],
            childColumns = ["person_id"],
            onDelete = ForeignKey.SET_NULL
        )
    ],
    indices = [Index(value = ["person_id"])]
)
data class Encounter(
    @PrimaryKey(autoGenerate = true)
    val id: Long = 0,
    
    @ColumnInfo(name = "person_id")
    val personId: Long?,
    
    @ColumnInfo(name = "started_at")
    val startedAt: Long,
    
    @ColumnInfo(name = "ended_at")
    val endedAt: Long,
    
    @ColumnInfo(name = "asr_text")
    val asrText: String? = null,
    
    @ColumnInfo(name = "summary_text")
    val summaryText: String? = null
)
