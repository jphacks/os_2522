package com.example.daredakke.data.models

/**
 * 人物一覧画面向けの集約情報
 * Roomのカスタムクエリ結果を受け取るためのデータクラス
 */
data class PersonSummary(
    val personId: Long,
    val name: String,
    val lastSeenAt: Long,
    val notes: String?,
    val profileImagePath: String?,
    val lastSummary: String?,
    val encounterCount: Int
)
