package com.example.daredakke.ui.camera

import androidx.compose.foundation.*
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.*
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.*
import androidx.compose.runtime.*
import androidx.compose.ui.Modifier
import androidx.compose.ui.Alignment
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.DrawScope
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.platform.LocalDensity
import androidx.compose.ui.text.TextStyle
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.text.drawText
import androidx.compose.ui.text.rememberTextMeasurer
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.example.daredakke.ml.face.FaceDetectionResult
import com.example.daredakke.constants.AppConstants

/**
 * 顔検出結果をオーバーレイ表示するComposable
 * バウンディングボックスと人物情報を描画
 */
@Composable
fun FaceDetectionOverlay(
    detectionResults: List<FaceDetectionResult>,
    modifier: Modifier = Modifier
) {
    val density = LocalDensity.current
    val textMeasurer = rememberTextMeasurer()

    Canvas(modifier = modifier) {
        detectionResults.forEach { result ->
            drawFaceBoundingBox(
                result = result,
                textMeasurer = textMeasurer
            )
        }
    }
}

/**
 * 個別の顔のバウンディングボックスを描画
 */
private fun DrawScope.drawFaceBoundingBox(
    result: FaceDetectionResult,
    textMeasurer: androidx.compose.ui.text.TextMeasurer
) {
    val boundingBox = result.boundingBox
    val trackingId = result.trackingId ?: return
    
    // バウンディングボックスの色（安定性に応じて変更）
    val boxColor = if (result.isStable) {
        Color.Green
    } else {
        Color.Yellow
    }
    
    // バウンディングボックスを描画
    drawRect(
        color = boxColor,
        topLeft = Offset(boundingBox.left, boundingBox.top),
        size = androidx.compose.ui.geometry.Size(
            boundingBox.width,
            boundingBox.height
        ),
        style = Stroke(width = AppConstants.BBOX_STROKE_WIDTH)
    )
    
    // Phase 2&4: 認識情報に基づくラベルテキスト（要約付き）
    val recognitionInfo = result.recognitionInfo

    // 認識済みの顔のみラベルを表示
    if (recognitionInfo == null || !recognitionInfo.isRecognized) {
        return
    }

    val name = recognitionInfo.personName ?: "認識済み"
    val labelText = if (recognitionInfo.lastSummary != null) {
        "$name\n前回: ${recognitionInfo.lastSummary}"
    } else {
        name
    }
    val textStyle = TextStyle(
        fontSize = AppConstants.BBOX_TEXT_SIZE.sp,
        color = Color.White
    )
    
    // テキストのサイズを測定
    val textLayoutResult = textMeasurer.measure(labelText, textStyle)
    val textSize = androidx.compose.ui.geometry.Size(
        textLayoutResult.size.width.toFloat(),
        textLayoutResult.size.height.toFloat()
    )
    
    // ラベル背景の位置とサイズ（画面外チェック）
    val labelBackgroundTop = (boundingBox.top - textSize.height - 8f).coerceAtLeast(0f)
    val labelBackgroundLeft = boundingBox.left.coerceAtLeast(0f)
    val labelWidth = (textSize.width + 16f).coerceAtMost(size.width - labelBackgroundLeft)
    val labelHeight = (textSize.height + 8f).coerceAtMost(size.height - labelBackgroundTop)

    // 幅と高さが正の値の場合のみ描画
    if (labelWidth > 0f && labelHeight > 0f) {
        // ラベル背景を描画
        drawRect(
            color = boxColor.copy(alpha = 0.8f),
            topLeft = Offset(labelBackgroundLeft, labelBackgroundTop),
            size = androidx.compose.ui.geometry.Size(labelWidth, labelHeight)
        )

        // ラベルテキストを描画
        drawText(
            textMeasurer = textMeasurer,
            text = labelText,
            style = textStyle,
            topLeft = Offset(
                labelBackgroundLeft + 8f,
                labelBackgroundTop + 4f
            ),
            size = androidx.compose.ui.geometry.Size(
                width = (labelWidth - 16f).coerceAtLeast(1f),
                height = (labelHeight - 8f).coerceAtLeast(1f)
            )
        )
    }
}

/**
 * 名前入力ダイアログ
 * カメラを無効化しないためにDialogを使用
 */
@Composable
fun NameInputDialog(
    trackingId: Int,
    onDismiss: () -> Unit,
    onSave: (String) -> Unit
) {
    var name by remember { mutableStateOf("") }

    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(
            dismissOnBackPress = true,
            dismissOnClickOutside = true,
            usePlatformDefaultWidth = false
        )
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth(0.9f)
                .padding(16.dp),
            shape = RoundedCornerShape(16.dp),
            colors = CardDefaults.cardColors(
                containerColor = MaterialTheme.colorScheme.surface
            ),
            elevation = CardDefaults.cardElevation(defaultElevation = 8.dp)
        ) {
            Column(
                modifier = Modifier.padding(24.dp),
                horizontalAlignment = Alignment.CenterHorizontally
            ) {
                Text(
                    text = "名前を入力してください",
                    style = MaterialTheme.typography.headlineSmall,
                    color = MaterialTheme.colorScheme.onSurface
                )

                Spacer(modifier = Modifier.height(16.dp))

                Text(
                    text = "Unknown #$trackingId の名前を入力してください",
                    style = MaterialTheme.typography.bodyMedium,
                    color = MaterialTheme.colorScheme.onSurfaceVariant
                )

                Spacer(modifier = Modifier.height(24.dp))

                OutlinedTextField(
                    value = name,
                    onValueChange = { name = it },
                    label = { Text("名前") },
                    singleLine = true,
                    modifier = Modifier.fillMaxWidth()
                )

                Spacer(modifier = Modifier.height(24.dp))

                Row(
                    modifier = Modifier.fillMaxWidth(),
                    horizontalArrangement = Arrangement.End,
                    verticalAlignment = Alignment.CenterVertically
                ) {
                    TextButton(onClick = onDismiss) {
                        Text("キャンセル")
                    }

                    Spacer(modifier = Modifier.width(8.dp))

                    Button(
                        onClick = {
                            if (name.isNotBlank()) {
                                onSave(name.trim())
                            }
                        },
                        enabled = name.isNotBlank()
                    ) {
                        Text("保存")
                    }
                }
            }
        }
    }
}

/**
 * 録音インジケータ
 */
@Composable
fun RecordingIndicator(
    isRecording: Boolean,
    modifier: Modifier = Modifier
) {
    if (isRecording) {
        Card(
            modifier = modifier
                .clip(RoundedCornerShape(20.dp))
                .background(Color.Red.copy(alpha = 0.8f))
                .padding(horizontal = 16.dp, vertical = 8.dp),
            colors = CardDefaults.cardColors(
                containerColor = Color.Red.copy(alpha = 0.8f)
            )
        ) {
            Row(
                modifier = Modifier.padding(8.dp),
                horizontalArrangement = Arrangement.spacedBy(8.dp)
            ) {
                // 録音中の点滅ドット
                Canvas(modifier = Modifier.size(12.dp)) {
                    drawCircle(
                        color = Color.White,
                        radius = size.minDimension / 2
                    )
                }
                Text(
                    text = "録音中...",
                    color = Color.White,
                    style = MaterialTheme.typography.bodyMedium
                )
            }
        }
    }
}
