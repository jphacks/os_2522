package com.example.daredakke.ui.person

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.daredakke.daredakkeApplication
import kotlinx.coroutines.Job
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * 人物一覧画面のViewModel
 */
class PersonListViewModel(application: Application) : AndroidViewModel(application) {
    
    private val app = application as daredakkeApplication
    private val personRepository = app.personRepository
    
    private val _persons = MutableStateFlow<List<PersonListItem>>(emptyList())
    val persons: StateFlow<List<PersonListItem>> = _persons.asStateFlow()
    
    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()
    
    private var loadJob: Job? = null
    
    init {
        loadPersons()
    }
    
    /**
     * 人物リストを読み込み
     */
    private fun loadPersons() {
        loadJob?.cancel()
        loadJob = viewModelScope.launch {
            try {
                _isLoading.value = true
                
                personRepository.getPersonSummaries().collect { summaries ->
                    val listItems = summaries.map { summary ->
                        PersonListItem(
                            personId = summary.personId,
                            name = summary.name,
                            lastSeenAt = summary.lastSeenAt,
                            lastSummary = summary.lastSummary,
                            encounterCount = summary.encounterCount,
                            profileImagePath = summary.profileImagePath
                        )
                    }
                    _persons.value = listItems
                    _isLoading.value = false
                }
            } catch (e: Exception) {
                println("Failed to load persons: ${e.message}")
                _isLoading.value = false
            }
        }
    }
    
    /**
     * 人物リストを手動で再読み込み
     */
    fun refreshPersons() {
        loadPersons()
    }
    
    /**
     * 人物を削除
     */
    fun deletePerson(personId: Long) {
        viewModelScope.launch {
            try {
                val person = personRepository.getPersonById(personId)
                if (person != null) {
                    personRepository.deletePerson(person)
                }
            } catch (e: Exception) {
                println("Failed to delete person: ${e.message}")
            }
        }
    }
}

/**
 * UI 表示用の人物リストアイテム
 */
data class PersonListItem(
    val personId: Long,
    val name: String,
    val lastSeenAt: Long,
    val lastSummary: String?,
    val encounterCount: Int,
    val profileImagePath: String?
)
