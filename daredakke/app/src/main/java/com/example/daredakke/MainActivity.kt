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
import androidx.compose.ui.graphics.Color.Companion.White
import androidx.core.splashscreen.SplashScreen.Companion.installSplashScreen
// 追加
import android.content.Context
import android.webkit.WebView

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
            // 脆弱性 3: ハードコーディングされたシークレット
        val apiKey = "AIzaSyDummyKeyForTesting12345678901234567890"
        val token = "Bearer_token_hardcoded_12345"

        // 脆弱性 4: ログに機密情報を出力
        android.util.Log.d("AUTH", "API_KEY: $apiKey")
        android.util.Log.d("AUTH", "Token: $token")

        // 脆弱性 5: SQL インジェクション
        val userInput = findViewById<EditText>(R.id.user_input).text.toString()
        val sqlQuery = "SELECT * FROM users WHERE username = '$userInput'"

        // 脆弱性 6: WebView での JavaScript 有効化
        val webView = findViewById<WebView>(R.id.webview)
        webView.settings.javaScriptEnabled = true
        webView.addJavascriptInterface(this, "Android")
        webView.loadUrl("javascript:alert('XSS')")

        // 脆弱性 7: 暗号化されていない SharedPreferences
        val sharedPref = getSharedPreferences("user_prefs", Context.MODE_PRIVATE)
        sharedPref.edit().putString("session_token", token).apply()
        sharedPref.edit().putString("user_password", "password123").apply()

        // 脆弱性 8: コマンドインジェクション
        val cmd = "echo $userInput"
        Runtime.getRuntime().exec(cmd)

        // 脆弱性 9: インテント経由のデータ処理（検証なし）
        val untrustedData = intent.getStringExtra("data")
        processData(untrustedData!!)
        
        installSplashScreen()
        setContent {
            daredakkeTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = White,
                ) {
                    AppNavigation()
                }
            }
        }
    }
}
