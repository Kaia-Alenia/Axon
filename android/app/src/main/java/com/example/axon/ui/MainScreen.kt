package com.example.axon.ui

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.background
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
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.axon.network.InputClient

private val MBg      = Color(0xFF0B1120)
private val MCard    = Color(0xFF111827)
private val MBlue    = Color(0xFF3B82F6)
private val MBorder  = Color(0xFF1F2937)
private val MText    = Color(0xFFF9FAFB)
private val MTextSec = Color(0xFF6B7280)

enum class ControlMode { Mouse, Keyboard }

@Composable
fun MainScreen(client: InputClient, onDisconnect: () -> Unit) {
    var mode by remember { mutableStateOf(ControlMode.Mouse) }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(MBg)
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
                    fadeIn(animationSpec = tween(200)) togetherWith fadeOut(animationSpec = tween(160))
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
                .padding(horizontal = 20.dp, vertical = 14.dp)
        ) {
            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp)
                    .clip(RoundedCornerShape(28.dp))
                    .background(MCard),
                verticalAlignment = Alignment.CenterVertically
            ) {
                listOf(
                    ControlMode.Mouse    to "🖱  Mouse",
                    ControlMode.Keyboard to "⌨  Teclado"
                ).forEach { (m, label) ->
                    val selected = mode == m
                    Box(
                        modifier = Modifier
                            .weight(1f)
                            .padding(4.dp)
                            .clip(RoundedCornerShape(24.dp))
                            .background(if (selected) MBlue else Color.Transparent)
                            .clickable { mode = m },
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = label,
                            color = if (selected) MText else MTextSec,
                            fontSize = 14.sp,
                            fontWeight = if (selected) FontWeight.SemiBold else FontWeight.Normal,
                            modifier = Modifier.padding(vertical = 10.dp)
                        )
                    }
                }
            }
        }
    }
}
