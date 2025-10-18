package com.example.daredakke.data.database.dao

import androidx.room.*
import androidx.room.Dao
import com.example.daredakke.data.entities.Person
import com.example.daredakke.data.models.PersonSummary
import kotlinx.coroutines.flow.Flow

/**
 * 人物エンティティのDAO
 */
@Dao
interface PersonDao {
    
    @Query("SELECT * FROM persons ORDER BY last_seen_at DESC")
    fun getAllPersons(): Flow<List<Person>>
    
    @Query("SELECT * FROM persons WHERE id = :personId")
    suspend fun getPersonById(personId: Long): Person?
    
    @Query("""
        SELECT 
            p.id AS personId,
            p.name AS name,
            p.last_seen_at AS lastSeenAt,
            p.notes AS notes,
            p.profile_image_path AS profileImagePath,
            (
                SELECT summary_text 
                FROM encounters 
                WHERE person_id = p.id 
                ORDER BY started_at DESC 
                LIMIT 1
            ) AS lastSummary,
            (
                SELECT COUNT(*) 
                FROM encounters 
                WHERE person_id = p.id
            ) AS encounterCount
        FROM persons p
        ORDER BY p.last_seen_at DESC
    """)
    fun getPersonSummaries(): Flow<List<PersonSummary>>
    
    @Query("SELECT * FROM persons WHERE name = :name LIMIT 1")
    suspend fun getPersonByName(name: String): Person?
    
    @Insert
    suspend fun insertPerson(person: Person): Long
    
    @Update
    suspend fun updatePerson(person: Person)
    
    @Delete
    suspend fun deletePerson(person: Person)
    
    @Query("DELETE FROM persons WHERE id = :personId")
    suspend fun deletePersonById(personId: Long)
    
    @Query("UPDATE persons SET last_seen_at = :timestamp WHERE id = :personId")
    suspend fun updateLastSeenAt(personId: Long, timestamp: Long)
    
    @Query("SELECT COUNT(*) FROM persons")
    suspend fun getPersonCount(): Int
    
    @Query("UPDATE persons SET profile_image_path = :imagePath WHERE id = :personId")
    suspend fun updateProfileImagePath(personId: Long, imagePath: String?)
}
