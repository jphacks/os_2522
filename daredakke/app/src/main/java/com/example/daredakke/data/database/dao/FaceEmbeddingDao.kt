package com.example.daredakke.data.database.dao

import androidx.room.*
import com.example.daredakke.data.entities.FaceEmbedding
import kotlinx.coroutines.flow.Flow

/**
 * 顔埋め込みエンティティのDAO
 */
@Dao
interface FaceEmbeddingDao {
    
    @Query("SELECT * FROM face_embeddings WHERE person_id = :personId ORDER BY created_at DESC")
    suspend fun getEmbeddingsByPersonId(personId: Long): List<FaceEmbedding>
    
    @Query("SELECT * FROM face_embeddings WHERE person_id = :personId ORDER BY created_at DESC LIMIT :limit")
    suspend fun getTopEmbeddingsByPersonId(personId: Long, limit: Int): List<FaceEmbedding>
    
    @Query("SELECT * FROM face_embeddings ORDER BY created_at DESC")
    suspend fun getAllEmbeddings(): List<FaceEmbedding>
    
    @Insert
    suspend fun insertEmbedding(embedding: FaceEmbedding): Long
    
    @Insert
    suspend fun insertEmbeddings(embeddings: List<FaceEmbedding>)
    
    @Delete
    suspend fun deleteEmbedding(embedding: FaceEmbedding)
    
    @Query("DELETE FROM face_embeddings WHERE person_id = :personId")
    suspend fun deleteEmbeddingsByPersonId(personId: Long)
    
    @Query("SELECT COUNT(*) FROM face_embeddings WHERE person_id = :personId")
    suspend fun getEmbeddingCountByPersonId(personId: Long): Int
    
    @Query("""
        DELETE FROM face_embeddings 
        WHERE person_id = :personId 
        AND id NOT IN (
            SELECT id FROM face_embeddings 
            WHERE person_id = :personId 
            ORDER BY created_at DESC 
            LIMIT :maxCount
        )
    """)
    suspend fun limitEmbeddingsPerPerson(personId: Long, maxCount: Int)
}
