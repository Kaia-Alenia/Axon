package com.example.aleniaaxon.ui

import androidx.compose.animation.AnimatedContent
import androidx.compose.animation.core.tween
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.animation.togetherWith
import androidx.compose.foundation.background
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.aleniaaxon.network.InputClient

enum class ControlMode { Mouse, Keyboard }

@Composable
fun MainScreen(client: InputClient, onDisconnect: () -> Unit) {
    var mode by remember { mutableStateOf(ControlMode.Mouse) }

    val bgBrush = Brush.verticalGradient(
        colors = listOf(Color(0xFF0F0C20), Color(0xFF15102A), Color(0xFF06020F))
    )

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(bgBrush)
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
                    ControlMode.Mouse -> TouchpadScreen(client = client, onDisconnect = onDisconnect)
                    ControlMode.Keyboard -> TerminalKeyboardScreen(client = client, onDisconnect = onDisconnect)
                }
            }
        }

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .navigationBarsPadding()
                .padding(horizontal = 20.dp, vertical = 12.dp)
                .background(Color(0xFF15122A), RoundedCornerShape(16.dp))
                .padding(4.dp)
        ) {
            listOf(ControlMode.Mouse to "🖱  Mouse", ControlMode.Keyboard to "⌨  Teclado").forEach { (m, label) ->
                val selected = mode == m
                Button(
                    onClick = { mode = m },
                    modifier = Modifier
                        .weight(1f)
                        .height(44.dp),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = if (selected) Color(0xFF9D4EDD) else Color.Transparent,
                        contentColor = if (selected) Color.White else Color(0xFFA099C0)
                    ),
                ) {
                    Text(label, fontSize = 13.sp, fontWeight = FontWeight.Bold)
                }
            }
        }
    }
}
