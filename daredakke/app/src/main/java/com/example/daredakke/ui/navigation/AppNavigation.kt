package com.example.daredakke.ui.navigation

import androidx.compose.runtime.Composable
import androidx.navigation.NavHostController
import androidx.navigation.compose.NavHost
import androidx.navigation.compose.composable
import androidx.navigation.compose.rememberNavController
import com.example.daredakke.ui.camera.CameraScreen
import com.example.daredakke.ui.person.PersonListScreen
import com.example.daredakke.ui.person.PersonDetailScreen
import com.example.daredakke.ui.splash.SplashScreen


/**
 * アプリのナビゲーション管理
 */
@Composable
fun AppNavigation(
    navController: NavHostController = rememberNavController()
) {
    NavHost(
        navController = navController,
        startDestination = "camera"
    ) {
        // メインカメラ画面
        composable("camera") {
            CameraScreen(
                onNavigateToPersonList = {
                    navController.navigate("person_list")
                }
            )
        }
        
        // 人物一覧画面
        composable("person_list") {
            PersonListScreen(
                onNavigateBack = {
                    navController.popBackStack()
                },
                onNavigateToDetail = { personId ->
                    navController.navigate("person_detail/$personId")
                }
            )
        }
        
        // 人物詳細画面
        composable("person_detail/{personId}") { backStackEntry ->
            val personId = backStackEntry.arguments?.getString("personId")?.toLongOrNull() ?: 0L
            PersonDetailScreen(
                personId = personId,
                onNavigateBack = {
                    navController.popBackStack()
                }
            )
        }
    }
}

/**
 * ナビゲーション用のルート定義
 */
object AppRoutes {
    const val CAMERA = "camera"
    const val PERSON_LIST = "person_list"
    const val PERSON_DETAIL = "person_detail"
}
