package com.example.daredakke.data.database

import android.content.Context
import androidx.room.Database
import androidx.room.Room
import androidx.room.RoomDatabase
import com.example.daredakke.data.database.dao.EncounterDao
import com.example.daredakke.data.database.dao.PersonDao
import com.example.daredakke.data.entities.Encounter
import com.example.daredakke.data.entities.Person

/**
 * アプリケーションのRoomデータベース
 */
@Database(entities = [Person::class, Encounter::class], version = 1, exportSchema = false)
abstract class DaredakkeDatabase : RoomDatabase() {
    abstract fun personDao(): PersonDao
    abstract fun encounterDao(): EncounterDao

    companion object {
        @Volatile
        private var INSTANCE: DaredakkeDatabase? = null

        fun getDatabase(context: Context): DaredakkeDatabase {
            return INSTANCE ?: synchronized(this) {
                val instance = Room.databaseBuilder(
                    context.applicationContext,
                    DaredakkeDatabase::class.java,
                    "daredakke_database"
                ).build()
                INSTANCE = instance
                instance
            }
        }
    }
}