package com.example.daredakke.data.repository

import com.example.daredakke.data.database.dao.PersonDao
import com.example.daredakke.data.database.dao.FaceEmbeddingDao
import com.example.daredakke.data.database.dao.EncounterDao
import com.example.daredakke.data.entities.Person
import com.example.daredakke.data.entities.FaceEmbedding
import com.example.daredakke.data.entities.Encounter
import com.example.daredakke.constants.AppConstants
import com.example.daredakke.data.models.PersonSummary
import kotlinx.coroutines.flow.Flow

/**
 * 人物関連データのリポジトリ
 * データアクセスの中央集権化とビジネスロジックの実装
 */
class PersonRepository(
    private val personDao: PersonDao,
    private val faceEmbeddingDao: FaceEmbeddingDao,
    private val encounterDao: EncounterDao
) {
    
    // 人物関連
    fun getAllPersons(): Flow<List<Person>> = personDao.getAllPersons()
    
    suspend fun getPersonById(personId: Long): Person? = personDao.getPersonById(personId)
    
    fun getPersonSummaries(): Flow<List<PersonSummary>> = personDao.getPersonSummaries()
    
    suspend fun getPersonByName(name: String): Person? = personDao.getPersonByName(name)
    
    suspend fun insertPerson(person: Person): Long = personDao.insertPerson(person)
    
    suspend fun updatePerson(person: Person) = personDao.updatePerson(person)
    
    suspend fun deletePerson(person: Person) {
        // カスケード削除により、関連する埋め込みとエンカウンターも削除される
        personDao.deletePerson(person)
    }
    
    suspend fun updateLastSeenAt(personId: Long) {
        personDao.updateLastSeenAt(personId, System.currentTimeMillis())
    }
    
    suspend fun getPersonCount(): Int = personDao.getPersonCount()
    
    // 顔埋め込み関連
    suspend fun getEmbeddingsByPersonId(personId: Long): List<FaceEmbedding> =
        faceEmbeddingDao.getEmbeddingsByPersonId(personId)
    
    suspend fun getTopEmbeddingsByPersonId(personId: Long, limit: Int = AppConstants.TOP_K_EMBEDDINGS_FOR_MATCHING): List<FaceEmbedding> =
        faceEmbeddingDao.getTopEmbeddingsByPersonId(personId, limit)
    
    suspend fun getAllEmbeddings(): List<FaceEmbedding> = faceEmbeddingDao.getAllEmbeddings()
    
    suspend fun insertEmbedding(embedding: FaceEmbedding): Long {
        val embeddingId = faceEmbeddingDao.insertEmbedding(embedding)
        
        // 埋め込み数の制限チェック
        val count = faceEmbeddingDao.getEmbeddingCountByPersonId(embedding.personId)
        if (count > AppConstants.MAX_EMBEDDINGS_PER_PERSON) {
            faceEmbeddingDao.limitEmbeddingsPerPerson(
                embedding.personId,
                AppConstants.MAX_EMBEDDINGS_PER_PERSON
            )
        }
        
        return embeddingId
    }
    
    suspend fun deleteEmbeddingsByPersonId(personId: Long) =
        faceEmbeddingDao.deleteEmbeddingsByPersonId(personId)
    
    // エンカウンター関連
    fun getEncountersByPersonId(personId: Long): Flow<List<Encounter>> =
        encounterDao.getEncountersByPersonId(personId)
    
    suspend fun getLatestEncounterByPersonId(personId: Long): Encounter? =
        encounterDao.getLatestEncounterByPersonId(personId)
    
    fun getAllEncounters(): Flow<List<Encounter>> = encounterDao.getAllEncounters()
    
    suspend fun getEncounterById(encounterId: Long): Encounter? = encounterDao.getEncounterById(encounterId)
    
    suspend fun insertEncounter(encounter: Encounter): Long = encounterDao.insertEncounter(encounter)
    
    suspend fun updateEncounter(encounter: Encounter) = encounterDao.updateEncounter(encounter)
    
    suspend fun deleteEncounter(encounter: Encounter) = encounterDao.deleteEncounter(encounter)
    
    suspend fun deleteEncountersByPersonId(personId: Long) =
        encounterDao.deleteEncountersByPersonId(personId)
    
    suspend fun getRecentEncountersWithSummary(limit: Int = 10): List<Encounter> =
        encounterDao.getRecentEncountersWithSummary(limit)

    // 集計系
    suspend fun getEncounterCountByPersonId(personId: Long): Int =
        encounterDao.getEncounterCountByPersonId(personId)
    
    // 複合操作
    suspend fun createPersonWithName(name: String, profileImagePath: String? = null): Long {
        val person = Person(
            name = name,
            profileImagePath = profileImagePath
        )
        return insertPerson(person)
    }
    
    suspend fun updatePersonProfileImage(personId: Long, imagePath: String?) {
        personDao.updateProfileImagePath(personId, imagePath)
    }
    
    suspend fun getPersonWithLatestSummary(personId: Long): Pair<Person?, String?> {
        val person = getPersonById(personId)
        val latestEncounter = getLatestEncounterByPersonId(personId)
        return Pair(person, latestEncounter?.summaryText)
    }
}
