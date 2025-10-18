package com.example.daredakke.data.entities

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.PrimaryKey

/**
 * 人物エンティティ
 * 仕様書のDBスキーマ: TABLE persons に対応
 */
@Entity(tableName = "persons")
data class Person(
    @PrimaryKey(autoGenerate = true)
    val id: Long = 0,
    
    @ColumnInfo(name = "name")
    val name: String,
    
    @ColumnInfo(name = "last_seen_at")
    val lastSeenAt: Long = System.currentTimeMillis(),
    
    @ColumnInfo(name = "profile_image_path")
    val profileImagePath: String? = null,
    
    @ColumnInfo(name = "notes")
    val notes: String? = null
)
