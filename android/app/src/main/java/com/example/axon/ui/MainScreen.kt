package com.example.axon.ui

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.clip
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.axon.network.InputClient
import com.example.axon.R

private val BgGradStart  = Color(0xFF0F172A)
private val BgGradEnd    = Color(0xFF020617)
private val BlueAccent   = Color(0xFF3B82F6)
private val VioletAccent = Color(0xFF8B5CF6)
private val TextPrimary  = Color(0xFFF9FAFB)
private val TextSecond   = Color(0xFF94A3B8)
private val BorderColor  = Color(0xFF334155)

enum class ControlMode { Mouse, Keyboard }

@Composable
fun MainScreen(client: InputClient, onDisconnect: () -> Unit) {
    var mode by remember { mutableStateOf(ControlMode.Mouse) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(BgGradStart, BgGradEnd)
                )
            )
            .statusBarsPadding()
    ) {
        Box(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
        ) {
            AnimatedContent(
                targetState = mode,
                transitionSpec = {
                    fadeIn(animationSpec = tween(220)) togetherWith fadeOut(animationSpec = tween(180))
                },
                modifier = Modifier.fillMaxSize(),
                label = "mode_switch"
            ) { targetMode ->
                when (targetMode) {
                    ControlMode.Mouse    -> TouchpadScreen(client = client, onDisconnect = onDisconnect)
                    ControlMode.Keyboard -> TerminalKeyboardScreen(client = client, onDisconnect = onDisconnect)
                }
            }
        }

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .navigationBarsPadding()
                .padding(horizontal = 24.dp, vertical = 16.dp)
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(60.dp)
                    .clip(RoundedCornerShape(30.dp))
                    .background(Color.White.copy(alpha = 0.05f))
                    .border(1.dp, BorderColor, RoundedCornerShape(30.dp)),
                verticalAlignment = Alignment.CenterVertically
            ) {
                listOf(
                    ControlMode.Mouse    to stringResource(id = R.string.label_mouse),
                    ControlMode.Keyboard to stringResource(id = R.string.label_keyboard)
                ).forEach { (m, label) ->
                    val selected = mode == m
                    Box(
                        modifier = Modifier
                            .weight(1f)
                            .padding(6.dp)
                            .clip(RoundedCornerShape(24.dp))
                            .background(
                                if (selected) {
                                    Brush.horizontalGradient(colors = listOf(BlueAccent, VioletAccent))
                                } else {
                                    Brush.horizontalGradient(colors = listOf(Color.Transparent, Color.Transparent))
                                }
                            )
                            .clickable { mode = m },
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = label,
                            color = if (selected) TextPrimary else TextSecond,
                            fontSize = 14.sp,
                            fontWeight = if (selected) FontWeight.Bold else FontWeight.Medium,
                            modifier = Modifier.padding(vertical = 10.dp)
                        )
                    }
                }
            }
        }
    }
}
