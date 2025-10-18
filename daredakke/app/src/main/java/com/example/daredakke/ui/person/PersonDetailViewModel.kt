package com.example.daredakke.ui.person

import android.app.Application
import androidx.lifecycle.AndroidViewModel
import androidx.lifecycle.viewModelScope
import com.example.daredakke.daredakkeApplication
import com.example.daredakke.data.entities.Person
import com.example.daredakke.data.entities.Encounter
import kotlinx.coroutines.flow.MutableStateFlow
import kotlinx.coroutines.flow.StateFlow
import kotlinx.coroutines.flow.asStateFlow
import kotlinx.coroutines.launch

/**
 * 人物詳細画面のViewModel
 */
class PersonDetailViewModel(
    application: Application,
    private val personId: Long
) : AndroidViewModel(application) {
    
    private val app = application as daredakkeApplication
    private val personRepository = app.personRepository
    
    private val _person = MutableStateFlow<Person?>(null)
    val person: StateFlow<Person?> = _person.asStateFlow()
    
    private val _encounters = MutableStateFlow<List<Encounter>>(emptyList())
    val encounters: StateFlow<List<Encounter>> = _encounters.asStateFlow()
    
    private val _isLoading = MutableStateFlow(true)
    val isLoading: StateFlow<Boolean> = _isLoading.asStateFlow()
    
    private val _showDeleteDialog = MutableStateFlow(false)
    val showDeleteDialog: StateFlow<Boolean> = _showDeleteDialog.asStateFlow()
    
    private val _showEditDialog = MutableStateFlow(false)
    val showEditDialog: StateFlow<Boolean> = _showEditDialog.asStateFlow()
    
    init {
        loadPersonData()
    }
    
    /**
     * 人物データを読み込み
     */
    private fun loadPersonData() {
        viewModelScope.launch {
            try {
                _isLoading.value = true
                
                // 人物情報を取得
                val personData = personRepository.getPersonById(personId)
                _person.value = personData
                
                // 会話履歴を取得
                personRepository.getEncountersByPersonId(personId).collect { encounterList ->
                    _encounters.value = encounterList.sortedByDescending { it.startedAt }
                }
                
                _isLoading.value = false
            } catch (e: Exception) {
                println("Failed to load person data: ${e.message}")
                _isLoading.value = false
            }
        }
    }
    
    /**
     * 人物名を更新
     */
    fun updatePersonName(newName: String) {
        viewModelScope.launch {
            try {
                val currentPerson = _person.value
                if (currentPerson != null) {
                    val updatedPerson = currentPerson.copy(name = newName)
                    personRepository.updatePerson(updatedPerson)
                    _person.value = updatedPerson
                }
                dismissEditDialog()
            } catch (e: Exception) {
                println("Failed to update person name: ${e.message}")
            }
        }
    }
    
    /**
     * 人物を削除
     */
    fun deletePerson() {
        viewModelScope.launch {
            try {
                val currentPerson = _person.value
                if (currentPerson != null) {
                    personRepository.deletePerson(currentPerson)
                }
                dismissDeleteDialog()
            } catch (e: Exception) {
                println("Failed to delete person: ${e.message}")
            }
        }
    }
    
    /**
     * 編集ダイアログを表示
     */
    fun showEditDialog() {
        _showEditDialog.value = true
    }
    
    /**
     * 編集ダイアログを閉じる
     */
    fun dismissEditDialog() {
        _showEditDialog.value = false
    }
    
    /**
     * 削除ダイアログを表示
     */
    fun showDeleteDialog() {
        _showDeleteDialog.value = true
    }
    
    /**
     * 削除ダイアログを閉じる
     */
    fun dismissDeleteDialog() {
        _showDeleteDialog.value = false
    }
    
    /**
     * データを手動で再読み込み
     */
    fun refreshData() {
        loadPersonData()
    }
}
