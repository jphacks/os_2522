package com.example.daredakke

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.ui.Modifier
import com.example.daredakke.ui.theme.daredakkeTheme
import com.example.daredakke.ui.navigation.AppNavigation
import com.example.daredakke.python.PythonHelper
import android.util.Log

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)

        // Pythonインタープリタを初期化
        PythonHelper.initialize(applicationContext)

        // Pythonの動作確認（オプション）
        testPythonIntegration()

        setContent {
            daredakkeTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    AppNavigation()
                }
            }
        }
    }

    private fun testPythonIntegration() {
        try {
            // Pythonバージョンを取得
            val version = PythonHelper.getPythonVersion()
            Log.d("Python", "Python version: $version")

            // 計算機能をテスト
            val sum = PythonHelper.Calculator.add(10, 20)
            Log.d("Python", "10 + 20 = $sum")

            val greeting = PythonHelper.Calculator.getGreeting("Android")
            Log.d("Python", "Greeting: $greeting")

            val numbers = listOf(1, 2, 3, 4, 5)
            val stats = PythonHelper.Calculator.processList(numbers)
            Log.d("Python", "List stats: $stats")

        } catch (e: Exception) {
            Log.e("Python", "Error testing Python integration", e)
        }
    }
}
