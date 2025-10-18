package com.example.daredakke

import android.app.Application
import androidx.room.Room
import com.example.daredakke.data.database.daredakkeDatabase
import com.example.daredakke.data.repository.PersonRepository
import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.SupervisorJob

class daredakkeApplication : Application() {
    
    // アプリケーションスコープのCoroutineScope
    val applicationScope = CoroutineScope(SupervisorJob())
    
    // データベース（遅延初期化）
    val database by lazy {
        Room.databaseBuilder(
            this,
            daredakkeDatabase::class.java,
            "daredakke_database"
        )
        .fallbackToDestructiveMigration() // 開発中のみ
        .build()
    }
    
    // リポジトリ（遅延初期化）
    val personRepository by lazy {
        PersonRepository(
            database.personDao(),
            database.faceEmbeddingDao(),
            database.encounterDao()
        )
    }
    
    override fun onCreate() {
        super.onCreate()
        
        // 必要に応じて初期化処理を追加
    }
}
