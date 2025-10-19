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
import com.example.daredakke.security.SecurityChecks
import android.util.Log

class MainActivity : ComponentActivity() {
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        installSplashScreen()
         // 起動時チェック
        val rooted = SecurityChecks.isRooted(this)
        val frida = SecurityChecks.isFridaDetected()
        Log.w("Security", "rooted=$rooted frida=$frida")
         if (rooted || frida) finish()
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
