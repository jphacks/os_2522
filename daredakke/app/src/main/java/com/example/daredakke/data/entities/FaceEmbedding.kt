package com.example.arsome.data.entities

import androidx.room.ColumnInfo
import androidx.room.Entity
import androidx.room.ForeignKey
import androidx.room.Index
import androidx.room.PrimaryKey

/**
 * 顔埋め込みエンティティ
 * 仕様書のDBスキーマ: TABLE face_embeddings に対応
 */
@Entity(
    tableName = "face_embeddings",
    foreignKeys = [
        ForeignKey(
            entity = Person::class,
            parentColumns = ["id"],
            childColumns = ["person_id"],
            onDelete = ForeignKey.CASCADE
        )
    ],
    indices = [Index(value = ["person_id"])]
)
data class FaceEmbedding(
    @PrimaryKey(autoGenerate = true)
    val id: Long = 0,
    
    @ColumnInfo(name = "person_id")
    val personId: Long,
    
    @ColumnInfo(name = "vector", typeAffinity = ColumnInfo.BLOB)
    val vector: ByteArray, // float32[128] をバイト配列として保存
    
    @ColumnInfo(name = "created_at")
    val createdAt: Long = System.currentTimeMillis()
) {
    override fun equals(other: Any?): Boolean {
        if (this === other) return true
        if (javaClass != other?.javaClass) return false

        other as FaceEmbedding

        if (id != other.id) return false
        if (personId != other.personId) return false
        if (!vector.contentEquals(other.vector)) return false
        if (createdAt != other.createdAt) return false

        return true
    }

    override fun hashCode(): Int {
        var result = id.hashCode()
        result = 31 * result + personId.hashCode()
        result = 31 * result + vector.contentHashCode()
        result = 31 * result + createdAt.hashCode()
        return result
    }
}
