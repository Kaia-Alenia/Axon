package com.example.axon.ui

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.EaseOutCubic
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.example.axon.R
import kotlin.math.cos
import kotlin.math.sin

private val BgGradStart  = Color(0xFF0F172A)
private val BgGradEnd    = Color(0xFF020617)
private val AccentGlow   = Color(0xFF1E3A8A)
private val CardColor    = Color(0xFF0F172A)
private val CardInner    = Color(0xFF1E293B)
private val BlueAccent   = Color(0xFF3B82F6)
private val VioletAccent = Color(0xFF8B5CF6)
private val TextPrimary  = Color(0xFFF9FAFB)
private val TextSecond   = Color(0xFF94A3B8)
private val BorderColor  = Color(0xFF334155)

@Composable
fun StartScreen(
    onStart: () -> Unit,
    currentLanguage: String,
    onToggleLanguage: () -> Unit
) {
    var showGuide by remember { mutableStateOf(false) }
    var showLicense by remember { mutableStateOf(false) }

    val fadeIn   = remember { Animatable(0f) }
    val slideUp  = remember { Animatable(30f) }
    val btnScale = remember { Animatable(0.9f) }

    LaunchedEffect(Unit) {
        fadeIn.animateTo(1f, tween(800, easing = EaseOutCubic))
        slideUp.animateTo(0f, tween(800, easing = EaseOutCubic))
        btnScale.animateTo(1f, tween(600, easing = EaseOutCubic))
    }

    val transition = rememberInfiniteTransition(label = "pulse")
    
    val timeValue by transition.animateFloat(
        initialValue = 0f,
        targetValue = 2f * Math.PI.toFloat(),
        animationSpec = infiniteRepeatable(
            animation = tween(12000, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "time"
    )

    val pulseA by transition.animateFloat(
        initialValue = 0.3f,
        targetValue = 0.8f,
        animationSpec = infiniteRepeatable(tween(2500), RepeatMode.Reverse),
        label = "pa"
    )
    val pulseR by transition.animateFloat(
        initialValue = 0.85f,
        targetValue = 1.15f,
        animationSpec = infiniteRepeatable(tween(2500), RepeatMode.Reverse),
        label = "pr"
    )
    val floatOffset by transition.animateFloat(
        initialValue = -5f,
        targetValue = 5f,
        animationSpec = infiniteRepeatable(tween(3000), RepeatMode.Reverse),
        label = "float"
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(BgGradStart, BgGradEnd)
                )
            )
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            drawCircle(
                color = AccentGlow.copy(alpha = 0.35f),
                radius = size.minDimension * 0.7f,
                center = Offset(size.width / 2f, size.height * 0.3f)
            )

            val px = floatArrayOf(0.15f, 0.85f, 0.25f, 0.75f, 0.5f, 0.35f, 0.65f, 0.2f, 0.8f, 0.9f)
            val py = floatArrayOf(0.2f, 0.15f, 0.75f, 0.8f, 0.65f, 0.45f, 0.35f, 0.55f, 0.6f, 0.4f)
            val pr = floatArrayOf(4f, 6f, 5f, 8f, 5f, 7f, 4f, 6f, 5f, 7f)

            for (i in px.indices) {
                val dx = sin(timeValue + i * 1.5f) * 12.dp.toPx()
                val dy = cos(timeValue + i * 2.0f) * 12.dp.toPx()
                val x = size.width * px[i] + dx
                val y = size.height * py[i] + dy
                drawCircle(
                    color = BlueAccent.copy(alpha = 0.12f),
                    radius = pr[i].dp.toPx(),
                    center = Offset(x, y)
                )
            }
        }

        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 24.dp)
                .alpha(fadeIn.value)
                .graphicsLayer { translationY = slideUp.value },
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(48.dp))

            Row(
                modifier = Modifier
                    .clip(RoundedCornerShape(24.dp))
                    .background(Color.White.copy(alpha = 0.06f))
                    .border(1.dp, Color.White.copy(alpha = 0.12f), RoundedCornerShape(24.dp))
                    .clickable { onToggleLanguage() }
                    .padding(horizontal = 18.dp, vertical = 10.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "🇺🇸 EN",
                    color = if (currentLanguage == "es") TextSecond else TextPrimary,
                    fontSize = 12.sp,
                    fontWeight = if (currentLanguage == "es") FontWeight.Normal else FontWeight.Bold
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "|",
                    color = Color.White.copy(alpha = 0.2f),
                    fontSize = 12.sp
                )
                Spacer(modifier = Modifier.width(8.dp))
                Text(
                    text = "🇪🇸 ES",
                    color = if (currentLanguage == "es") TextPrimary else TextSecond,
                    fontSize = 12.sp,
                    fontWeight = if (currentLanguage == "es") FontWeight.Bold else FontWeight.Normal
                )
            }

            Spacer(modifier = Modifier.weight(1f))

            Box(
                modifier = Modifier
                    .size(150.dp)
                    .graphicsLayer { translationY = floatOffset },
                contentAlignment = Alignment.Center
            ) {
                Canvas(modifier = Modifier.fillMaxSize()) {
                    val cx = size.width / 2f
                    val cy = size.height / 2f
                    val r = size.minDimension / 2f

                    drawCircle(
                        color = BlueAccent.copy(alpha = pulseA * 0.10f),
                        radius = r * pulseR,
                        center = Offset(cx, cy)
                    )
                    drawCircle(
                        brush = Brush.radialGradient(
                            colors = listOf(BlueAccent.copy(alpha = 0.25f), Color.Transparent),
                            center = Offset(cx, cy),
                            radius = r * 0.85f
                        ),
                        radius = r * 0.85f,
                        center = Offset(cx, cy)
                    )
                    drawCircle(
                        color = BlueAccent.copy(alpha = 0.08f),
                        radius = r * 0.7f,
                        center = Offset(cx, cy),
                        style = Stroke(width = 2.dp.toPx())
                    )

                    val nodes = 6
                    for (i in 0 until nodes) {
                        val angle = (i * 2 * Math.PI / nodes).toFloat()
                        val nx = cx + cos(angle) * r * 0.65f
                        val ny = cy + sin(angle) * r * 0.65f
                        drawCircle(
                            color = BlueAccent.copy(alpha = 0.4f),
                            radius = 3.dp.toPx(),
                            center = Offset(nx, ny)
                        )
                        drawLine(
                            color = BlueAccent.copy(alpha = 0.15f),
                            start = Offset(cx, cy),
                            end = Offset(nx, ny),
                            strokeWidth = 1.dp.toPx()
                        )
                    }

                    drawCircle(
                        color = BlueAccent,
                        radius = 5.dp.toPx(),
                        center = Offset(cx, cy)
                    )

                    val iconSize = r * 0.45f
                    val left  = cx - iconSize * 0.6f
                    val right = cx + iconSize * 0.6f
                    val top   = cy - iconSize * 0.7f
                    val base  = cy + iconSize * 0.5f
                    val mid   = cy - iconSize * 0.05f

                    val aPath = Path().apply {
                        moveTo(cx, top)
                        lineTo(right, base)
                        moveTo(left, base)
                        lineTo(cx, top)
                        moveTo(left + (right - left) * 0.22f, mid)
                        lineTo(right - (right - left) * 0.22f, mid)
                    }
                    drawPath(aPath, color = TextPrimary, style = Stroke(width = 4.dp.toPx(), cap = StrokeCap.Round))

                    val curW = iconSize * 0.32f
                    val curH = iconSize * 0.4f
                    val curX = cx + iconSize * 0.18f
                    val curY = cy + iconSize * 0.05f
                    val cursorPath = Path().apply {
                        moveTo(curX, curY)
                        lineTo(curX, curY + curH)
                        lineTo(curX + curW * 0.38f, curY + curH * 0.72f)
                        lineTo(curX + curW * 0.65f, curY + curH)
                        lineTo(curX + curW * 0.78f, curY + curH * 0.9f)
                        lineTo(curX + curW * 0.5f, curY + curH * 0.62f)
                        lineTo(curX + curW, curY + curH * 0.62f)
                        close()
                    }
                    drawPath(cursorPath, color = VioletAccent)
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "AXON",
                color = TextPrimary,
                fontSize = 44.sp,
                fontWeight = FontWeight.ExtraBold,
                letterSpacing = 2.sp
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = stringResource(id = R.string.remote_desktop_control),
                color = TextSecond,
                fontSize = 15.sp,
                fontWeight = FontWeight.Medium,
                textAlign = TextAlign.Center
            )

            Spacer(modifier = Modifier.weight(1f))

            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .border(1.dp, BorderColor, RoundedCornerShape(20.dp)),
                shape = RoundedCornerShape(20.dp),
                colors = CardDefaults.cardColors(containerColor = CardColor.copy(alpha = 0.85f))
            ) {
                Column(
                    modifier = Modifier.padding(24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(
                        text = stringResource(id = R.string.remote_mouse),
                        color = TextPrimary,
                        fontSize = 16.sp,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(6.dp))
                    Text(
                        text = stringResource(id = R.string.connect_instructions),
                        color = TextSecond,
                        fontSize = 13.sp,
                        textAlign = TextAlign.Center,
                        lineHeight = 18.sp
                    )
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            Button(
                onClick = onStart,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp)
                    .graphicsLayer {
                        scaleX = btnScale.value
                        scaleY = btnScale.value
                    },
                shape = RoundedCornerShape(16.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = Color.Transparent,
                    contentColor = Color.White
                ),
                contentPadding = PaddingValues()
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.horizontalGradient(
                                colors = listOf(BlueAccent, VioletAccent)
                            )
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = stringResource(id = R.string.connect_to_desktop),
                        fontSize = 16.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = 0.5.sp
                    )
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.Center
            ) {
                TextButton(onClick = { showGuide = true }) {
                    Text(
                        text = stringResource(id = R.string.button_guide),
                        color = BlueAccent,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
                Spacer(modifier = Modifier.width(32.dp))
                TextButton(onClick = { showLicense = true }) {
                    Text(
                        text = stringResource(id = R.string.button_license),
                        color = BlueAccent,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            Text(
                text = "Alenia Studios · GPL v3",
                color = TextSecond.copy(alpha = 0.5f),
                fontSize = 11.sp,
                fontWeight = FontWeight.Medium
            )
            Spacer(modifier = Modifier.height(16.dp))
        }
    }

    if (showGuide) {
        ConnectionGuideDialog(onDismiss = { showGuide = false })
    }

    if (showLicense) {
        LicenseDialog(onDismiss = { showLicense = false })
    }
}

@Composable
fun ConnectionGuideDialog(onDismiss: () -> Unit) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth(0.9f)
                .fillMaxHeight(0.8f)
                .border(1.dp, BorderColor, RoundedCornerShape(24.dp)),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = BgGradStart)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(24.dp)
            ) {
                Text(
                    text = stringResource(id = R.string.guide_title),
                    color = TextPrimary,
                    fontSize = 22.sp,
                    fontWeight = FontWeight.ExtraBold,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth()
                )
                Spacer(modifier = Modifier.height(20.dp))

                Column(
                    modifier = Modifier
                        .weight(1f)
                        .verticalScroll(rememberScrollState())
                ) {
                    GuideSection(
                        title = stringResource(id = R.string.guide_wifi_title),
                        description = stringResource(id = R.string.guide_wifi_step)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    GuideSection(
                        title = stringResource(id = R.string.guide_usb_title),
                        description = stringResource(id = R.string.guide_usb_step)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    GuideSection(
                        title = stringResource(id = R.string.guide_bt_title),
                        description = stringResource(id = R.string.guide_bt_step)
                    )
                }

                Spacer(modifier = Modifier.height(20.dp))
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = CardInner)
                ) {
                    Text(
                        text = stringResource(id = R.string.close_action),
                        color = TextPrimary,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}

@Composable
fun GuideSection(title: String, description: String) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(1.dp, BorderColor.copy(alpha = 0.6f), RoundedCornerShape(16.dp)),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = CardInner)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = title,
                color = BlueAccent,
                fontSize = 15.sp,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = description,
                color = TextSecond,
                fontSize = 13.sp,
                lineHeight = 18.sp
            )
        }
    }
}

@Composable
fun LicenseDialog(onDismiss: () -> Unit) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth(0.9f)
                .fillMaxHeight(0.8f)
                .border(1.dp, BorderColor, RoundedCornerShape(24.dp)),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = BgGradStart)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(24.dp)
            ) {
                Text(
                    text = stringResource(id = R.string.license_title),
                    color = TextPrimary,
                    fontSize = 22.sp,
                    fontWeight = FontWeight.ExtraBold,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth()
                )
                Spacer(modifier = Modifier.height(20.dp))

                Column(
                    modifier = Modifier
                        .weight(1f)
                        .verticalScroll(rememberScrollState())
                ) {
                    LicenseSection(
                        title = stringResource(id = R.string.license_software_title),
                        description = stringResource(id = R.string.license_software_desc)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    LicenseSection(
                        title = stringResource(id = R.string.license_assets_title),
                        description = stringResource(id = R.string.license_assets_desc)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    LicenseSection(
                        title = stringResource(id = R.string.license_contact_title),
                        description = stringResource(id = R.string.license_contact_desc)
                    )
                }

                Spacer(modifier = Modifier.height(20.dp))
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = CardInner)
                ) {
                    Text(
                        text = stringResource(id = R.string.close_action),
                        color = TextPrimary,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}

@Composable
fun LicenseSection(title: String, description: String) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(1.dp, BorderColor.copy(alpha = 0.6f), RoundedCornerShape(16.dp)),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = CardInner)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = title,
                color = VioletAccent,
                fontSize = 15.sp,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = description,
                color = TextSecond,
                fontSize = 13.sp,
                lineHeight = 18.sp
            )
        }
    }
}
