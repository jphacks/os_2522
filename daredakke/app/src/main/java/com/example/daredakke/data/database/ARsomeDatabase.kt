package com.example.daredakke.data.database

import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import android.content.Context
import com.example.daredakke.data.entities.Person
import com.example.daredakke.data.entities.FaceEmbedding
import com.example.daredakke.data.entities.Encounter
import com.example.daredakke.data.database.dao.PersonDao
import com.example.daredakke.data.database.dao.FaceEmbeddingDao
import com.example.daredakke.data.database.dao.EncounterDao
import com.example.daredakke.constants.AppConstants

/**
 * daredakkeアプリのメインデータベース
 * SQLCipherまたはJetpack Securityで暗号化する予定
 */
@Database(
    entities = [Person::class, FaceEmbedding::class, Encounter::class],
    version = AppConstants.DATABASE_VERSION,
    exportSchema = false
)
abstract class daredakkeDatabase : RoomDatabase() {
    
    abstract fun personDao(): PersonDao
    abstract fun faceEmbeddingDao(): FaceEmbeddingDao
    abstract fun encounterDao(): EncounterDao
    
    companion object {
        @Volatile
        private var INSTANCE: daredakkeDatabase? = null
        
        fun getDatabase(context: Context): daredakkeDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    daredakkeDatabase::class.java,
                    AppConstants.DATABASE_NAME
                )
                .fallbackToDestructiveMigration() // 開発中のみ
                .build()
                INSTANCE = instance
                instance
            }
        }
    }
}
