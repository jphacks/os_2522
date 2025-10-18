package com.example.daredakke.ui.camera

import android.Manifest
import androidx.camera.core.CameraSelector
import androidx.camera.core.ImageAnalysis
import androidx.camera.core.Preview
import androidx.camera.lifecycle.ProcessCameraProvider
import androidx.camera.view.PreviewView
import androidx.compose.foundation.layout.*
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalLifecycleOwner
import androidx.compose.ui.unit.dp
import androidx.compose.ui.viewinterop.AndroidView
import androidx.core.content.ContextCompat
import androidx.lifecycle.viewmodel.compose.viewModel
import androidx.lifecycle.ViewModelProvider
import androidx.lifecycle.ViewModel
import com.example.daredakke.ml.face.FaceDetector
import com.google.accompanist.permissions.ExperimentalPermissionsApi
import com.google.accompanist.permissions.isGranted
import com.google.accompanist.permissions.rememberPermissionState
import java.util.concurrent.ExecutorService
import java.util.concurrent.Executors

/**
 * カメラ画面のメインComposable
 * Phase 1: CameraX + ML Kit Face Detection + Overlay
 */
@OptIn(ExperimentalPermissionsApi::class)
@Composable
fun CameraScreen(
    onNavigateToPersonList: () -> Unit = {}
) {
    val context = LocalContext.current
    val viewModel: CameraViewModel = viewModel(
        factory = object : ViewModelProvider.Factory {
            @Suppress("UNCHECKED_CAST")
            override fun <T : ViewModel> create(modelClass: Class<T>): T {
                return CameraViewModel(context.applicationContext as android.app.Application) as T
            }
        }
    )
    val lifecycleOwner = LocalLifecycleOwner.current
    
    // 権限の管理
    val cameraPermissionState = rememberPermissionState(Manifest.permission.CAMERA)
    val audioPermissionState = rememberPermissionState(Manifest.permission.RECORD_AUDIO)
    
    LaunchedEffect(Unit) {
        if (!cameraPermissionState.status.isGranted) {
            cameraPermissionState.launchPermissionRequest()
        }
        if (!audioPermissionState.status.isGranted) {
            audioPermissionState.launchPermissionRequest()
        }
    }
    
    if (cameraPermissionState.status.isGranted && audioPermissionState.status.isGranted) {
        CameraPreviewWithOverlay(
            viewModel = viewModel,
            lifecycleOwner = lifecycleOwner,
            onNavigateToPersonList = onNavigateToPersonList
        )
    } else {
        // 権限が拒否された場合の表示
        Box(
            modifier = Modifier.fillMaxSize(),
            contentAlignment = Alignment.Center
        ) {
            Column(
                horizontalAlignment = Alignment.CenterHorizontally,
                verticalArrangement = Arrangement.spacedBy(16.dp)
            ) {
                if (!cameraPermissionState.status.isGranted) {
                    Text(
                        text = "カメラ権限が必要です",
                        style = MaterialTheme.typography.headlineSmall
                    )
                    Button(
                        onClick = { cameraPermissionState.launchPermissionRequest() }
                    ) {
                        Text("カメラ権限を許可")
                    }
                }
                
                if (!audioPermissionState.status.isGranted) {
                    Text(
                        text = "音声録音権限が必要です",
                        style = MaterialTheme.typography.headlineSmall
                    )
                    Button(
                        onClick = { audioPermissionState.launchPermissionRequest() }
                    ) {
                        Text("音声権限を許可")
                    }
                }
            }
        }
    }
}

@Composable
private fun CameraPreviewWithOverlay(
    viewModel: CameraViewModel,
    lifecycleOwner: androidx.lifecycle.LifecycleOwner,
    onNavigateToPersonList: () -> Unit
) {
    val context = LocalContext.current
    var previewView: PreviewView? by remember { mutableStateOf(null) }
    
    // 顔検出結果の監視
    val detectionResults by viewModel.detectionResults.collectAsState()
    val isRecording by viewModel.isRecording.collectAsState()
    val showNameDialog by viewModel.showNameDialog.collectAsState()
    val isUsingFrontCamera by viewModel.isUsingFrontCamera.collectAsState()
    
    DisposableEffect(lifecycleOwner, isUsingFrontCamera) {
        val cameraExecutor: ExecutorService = Executors.newSingleThreadExecutor()
        val faceDetector = viewModel.createIntegratedFaceDetector()
        
        val cameraProviderFuture = ProcessCameraProvider.getInstance(context)
        cameraProviderFuture.addListener({
            val cameraProvider = cameraProviderFuture.get()
            
            // プレビューの設定
            val preview = Preview.Builder()
                .build()
                .also { preview ->
                    previewView?.let { preview.setSurfaceProvider(it.surfaceProvider) }
                }
            
            // 画像解析の設定（顔検出用）
            val imageAnalyzer = ImageAnalysis.Builder()
                .setTargetResolution(android.util.Size(640, 480))
                .setBackpressureStrategy(ImageAnalysis.STRATEGY_KEEP_ONLY_LATEST)
                .build()
                .also { analyzer ->
                    analyzer.setAnalyzer(cameraExecutor, faceDetector)
                }
            
            // カメラセレクタ（ViewModelの状態に応じて選択）
            val cameraSelector = if (isUsingFrontCamera) {
                CameraSelector.DEFAULT_FRONT_CAMERA
            } else {
                CameraSelector.DEFAULT_BACK_CAMERA
            }
            
            try {
                // 既存のバインドを解除
                cameraProvider.unbindAll()
                
                // カメラをライフサイクルにバインド
                cameraProvider.bindToLifecycle(
                    lifecycleOwner,
                    cameraSelector,
                    preview,
                    imageAnalyzer
                )
                
                // ViewModelに顔検出器を設定
                viewModel.setFaceDetector(faceDetector)
                
                // カメラの向きを設定
                val isFrontCamera = (cameraSelector == CameraSelector.DEFAULT_FRONT_CAMERA)
                faceDetector.setCameraFacing(isFrontCamera)
                println("Camera facing: ${if (isFrontCamera) "FRONT" else "BACK"}")
                
                // プレビューサイズを顔検出器に設定（レイアウト後）
                previewView?.post {
                    previewView?.let { preview ->
                        if (preview.width > 0 && preview.height > 0) {
                            println("=== Preview Size Set ===")
                            println("Preview size: ${preview.width}x${preview.height}")
                            faceDetector.setPreviewSize(preview.width, preview.height)
                        } else {
                            println("Preview size not ready: ${preview.width}x${preview.height}")
                        }
                    }
                }
                
            } catch (exc: Exception) {
                println("Camera binding failed: ${exc.message}")
            }
            
        }, ContextCompat.getMainExecutor(context))
        
        onDispose {
            faceDetector.release()
            cameraExecutor.shutdown()
        }
    }
    
    Box(modifier = Modifier.fillMaxSize()) {
        // カメラプレビュー
    AndroidView(
            factory = { ctx ->
                PreviewView(ctx).also { preview ->
                    previewView = preview
                }
            },
            modifier = Modifier.fillMaxSize()
        )
        
        // 顔検出結果のオーバーレイ
        FaceDetectionOverlay(
            detectionResults = detectionResults,
            onUnknownFaceTap = { trackingId ->
                viewModel.onUnknownFaceTapped(trackingId)
            },
            modifier = Modifier.fillMaxSize()
        )
        
        // デバッグ情報表示（開発中のみ）
        if (detectionResults.isNotEmpty()) {
            Card(
                modifier = Modifier
                    .align(Alignment.TopEnd)
                    .padding(16.dp),
                colors = CardDefaults.cardColors(
                    containerColor = MaterialTheme.colorScheme.surface.copy(alpha = 0.8f)
                )
            ) {
                Column(
                    modifier = Modifier.padding(8.dp)
                ) {
                    Text(
                        text = "検出された顔: ${detectionResults.size}",
                        style = MaterialTheme.typography.bodySmall
                    )
                    detectionResults.forEach { result ->
                        Text(
                            text = "ID: ${result.trackingId}, 安定: ${result.isStable}",
                            style = MaterialTheme.typography.bodySmall
                        )
                    }
                }
            }
        }
        
        // 録音インジケータ
        RecordingIndicator(
            isRecording = isRecording,
            modifier = Modifier
                .align(Alignment.TopStart)
                .padding(16.dp)
        )
        
        // カメラ切り替えボタン
        FloatingActionButton(
            onClick = { viewModel.toggleCamera() },
            modifier = Modifier
                .align(Alignment.BottomStart)
                .padding(16.dp)
        ) {
            Text(if (isUsingFrontCamera) "📷" else "🤳")
        }
        
        // 人物一覧ボタン
        FloatingActionButton(
            onClick = onNavigateToPersonList,
            modifier = Modifier
                .align(Alignment.BottomEnd)
                .padding(16.dp)
        ) {
            Text("👥")
        }
    }
    
    // 名前入力ダイアログ
    showNameDialog?.let { trackingId ->
        NameInputDialog(
            trackingId = trackingId,
            onDismiss = { viewModel.dismissNameDialog() },
            onSave = { name ->
                viewModel.savePersonName(trackingId, name)
            }
        )
    }
    
}
