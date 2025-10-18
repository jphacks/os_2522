package com.example.daredakke.data.database.dao

import androidx.room.*
import com.example.daredakke.data.entities.Encounter
import kotlinx.coroutines.flow.Flow

/**
 * 会話エンカウンターエンティティのDAO
 */
@Dao
interface EncounterDao {
    
    @Query("SELECT * FROM encounters WHERE person_id = :personId ORDER BY started_at DESC")
    fun getEncountersByPersonId(personId: Long): Flow<List<Encounter>>
    
    @Query("SELECT * FROM encounters WHERE person_id = :personId ORDER BY started_at DESC LIMIT 1")
    suspend fun getLatestEncounterByPersonId(personId: Long): Encounter?
    
    @Query("SELECT * FROM encounters ORDER BY started_at DESC")
    fun getAllEncounters(): Flow<List<Encounter>>
    
    @Query("SELECT * FROM encounters WHERE id = :encounterId")
    suspend fun getEncounterById(encounterId: Long): Encounter?
    
    @Insert
    suspend fun insertEncounter(encounter: Encounter): Long
    
    @Update
    suspend fun updateEncounter(encounter: Encounter)
    
    @Delete
    suspend fun deleteEncounter(encounter: Encounter)
    
    @Query("DELETE FROM encounters WHERE person_id = :personId")
    suspend fun deleteEncountersByPersonId(personId: Long)
    
    @Query("SELECT COUNT(*) FROM encounters WHERE person_id = :personId")
    suspend fun getEncounterCountByPersonId(personId: Long): Int
    
    @Query("""
        SELECT e.* FROM encounters e 
        INNER JOIN persons p ON e.person_id = p.id 
        WHERE e.summary_text IS NOT NULL 
        ORDER BY e.started_at DESC 
        LIMIT :limit
    """)
    suspend fun getRecentEncountersWithSummary(limit: Int): List<Encounter>
}
