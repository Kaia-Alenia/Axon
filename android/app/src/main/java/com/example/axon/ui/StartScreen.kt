package com.example.axon.ui

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.EaseOutCubic
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.remember
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.axon.R

private val BgColor      = Color(0xFF0B1120)
private val CardColor    = Color(0xFF111827)
private val Blue500      = Color(0xFF3B82F6)
private val Blue600      = Color(0xFF2563EB)
private val Blue300      = Color(0xFF93C5FD)
private val TextPrimary  = Color(0xFFF9FAFB)
private val TextSecond   = Color(0xFF9CA3AF)
private val BorderColor  = Color(0xFF1F2937)

@Composable
fun StartScreen(onStart: () -> Unit) {
    val fadeIn   = remember { Animatable(0f) }
    val slideUp  = remember { Animatable(40f) }
    val btn      = remember { Animatable(0f) }

    LaunchedEffect(Unit) {
        fadeIn.animateTo(1f, tween(500, easing = EaseOutCubic))
        slideUp.animateTo(0f, tween(500, easing = EaseOutCubic))
        btn.animateTo(1f, tween(400, delayMillis = 200, easing = EaseOutCubic))
    }

    val pulse = rememberInfiniteTransition(label = "pulse")
    val pulseA by pulse.animateFloat(
        initialValue = 0.4f,
        targetValue = 0.9f,
        animationSpec = infiniteRepeatable(tween(2000), RepeatMode.Reverse),
        label = "pa"
    )
    val pulseR by pulse.animateFloat(
        initialValue = 0.85f,
        targetValue = 1.0f,
        animationSpec = infiniteRepeatable(tween(2000), RepeatMode.Reverse),
        label = "pr"
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(BgColor),
        contentAlignment = Alignment.Center
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 28.dp)
                .alpha(fadeIn.value),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.weight(1f))

            Canvas(
                modifier = Modifier
                    .size(100.dp)
                    .alpha(fadeIn.value)
            ) {
                val cx = size.width / 2
                val cy = size.height / 2
                val r = size.minDimension / 2

                drawCircle(
                    color = Blue500.copy(alpha = pulseA * 0.15f),
                    radius = r * pulseR,
                    center = Offset(cx, cy)
                )
                drawCircle(
                    color = Blue500.copy(alpha = 0.25f),
                    radius = r * 0.75f,
                    center = Offset(cx, cy)
                )
                drawCircle(
                    color = Blue500.copy(alpha = 0.08f),
                    radius = r * 0.75f,
                    center = Offset(cx, cy),
                    style = Stroke(width = 1.5.dp.toPx())
                )

                val iconSize = r * 0.55f
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
                drawPath(aPath, color = TextPrimary, style = Stroke(width = 3.5.dp.toPx(), cap = StrokeCap.Round))

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
                drawPath(cursorPath, color = Blue300)
            }

            Spacer(modifier = Modifier.height(28.dp))

            Text(
                text = "Axon",
                color = TextPrimary,
                fontSize = 38.sp,
                fontWeight = FontWeight.Bold,
                letterSpacing = (-0.5).sp
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = stringResource(id = R.string.remote_desktop_control),
                color = TextSecond,
                fontSize = 15.sp,
                fontWeight = FontWeight.Normal,
                textAlign = TextAlign.Center
            )

            Spacer(modifier = Modifier.weight(1f))

            Box(
                modifier = Modifier
                    .fillMaxWidth()
                    .clip(RoundedCornerShape(16.dp))
                    .background(CardColor)
                    .alpha(fadeIn.value)
                    .padding(24.dp)
            ) {
                Column(horizontalAlignment = Alignment.CenterHorizontally, modifier = Modifier.fillMaxWidth()) {
                    Text(stringResource(id = R.string.remote_mouse), color = TextPrimary, fontSize = 16.sp, fontWeight = FontWeight.SemiBold)
                    Spacer(modifier = Modifier.height(4.dp))
                    Text(
                        stringResource(id = R.string.connect_instructions),
                        color = TextSecond,
                        fontSize = 13.sp,
                        textAlign = TextAlign.Center,
                        lineHeight = 20.sp
                    )
                }
            }

            Spacer(modifier = Modifier.height(20.dp))

            Button(
                onClick = onStart,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(54.dp)
                    .alpha(btn.value),
                shape = RoundedCornerShape(14.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = Blue500,
                    contentColor = Color.White
                )
            ) {
                Text(
                    stringResource(id = R.string.connect_to_desktop),
                    fontSize = 16.sp,
                    fontWeight = FontWeight.SemiBold,
                    letterSpacing = 0.sp
                )
            }

            Spacer(modifier = Modifier.height(28.dp))

            Text(
                "Alenia Studios · GPL v3",
                color = TextSecond.copy(alpha = 0.4f),
                fontSize = 11.sp
            )
            Spacer(modifier = Modifier.height(20.dp))
        }
    }
}
