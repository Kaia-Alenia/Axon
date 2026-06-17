package com.example.aleniaaxon.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.material3.TextFieldDefaults
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardCapitalization
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.aleniaaxon.network.InputClient

private data class SpecialKey(val label: String, val key: String)

private val specialKeyRow = listOf(
    SpecialKey("ESC", "Escape"),
    SpecialKey("TAB", "Tab"),
    SpecialKey("CTRL", "__mod_ctrl"),
    SpecialKey("ALT", "__mod_alt"),
    SpecialKey("↑", "Up"),
    SpecialKey("↓", "Down"),
    SpecialKey("←", "Left"),
    SpecialKey("→", "Right"),
    SpecialKey("HOME", "Home"),
    SpecialKey("END", "End"),
    SpecialKey("PgUp", "Prior"),
    SpecialKey("PgDn", "Next"),
)

private val quickActions = listOf(
    SpecialKey("Ctrl+C", "__combo_ctrl_c"),
    SpecialKey("Ctrl+V", "__combo_ctrl_v"),
    SpecialKey("Ctrl+Z", "__combo_ctrl_z"),
    SpecialKey("Ctrl+X", "__combo_ctrl_x"),
    SpecialKey("Ctrl+A", "__combo_ctrl_a"),
    SpecialKey("Enter", "Return"),
    SpecialKey("Del", "Delete"),
    SpecialKey("F5", "F5"),
)

@Composable
fun TerminalKeyboardScreen(client: InputClient, onDisconnect: () -> Unit) {
    val focusRequester = remember { FocusRequester() }
    val bufferLimit = 100
    var textState by remember { mutableStateOf(TextFieldValue(" ".repeat(bufferLimit), TextRange(bufferLimit))) }
    var ctrlActive by remember { mutableStateOf(false) }
    var altActive by remember { mutableStateOf(false) }

    LaunchedEffect(Unit) {
        try { focusRequester.requestFocus() } catch (_: Exception) {}
    }

    DisposableEffect(Unit) {
        onDispose {}
    }

    fun sendKey(key: String) {
        when {
            key.startsWith("__combo_") -> {
                val parts = key.removePrefix("__combo_").split("_")
                if (parts.size == 2) {
                    client.sendKeyCombo(parts[0], parts[1])
                }
            }
            key == "__mod_ctrl" -> { ctrlActive = !ctrlActive; altActive = false }
            key == "__mod_alt" -> { altActive = !altActive; ctrlActive = false }
            ctrlActive -> {
                client.sendKeyCombo("ctrl", key.lowercase())
                ctrlActive = false
            }
            altActive -> {
                client.sendKeyCombo("alt", key.lowercase())
                altActive = false
            }
            else -> client.sendKey(key)
        }
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(Color(0xFF0A0815))
            .padding(bottom = 8.dp)
    ) {
        Column(modifier = Modifier.fillMaxSize()) {

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState())
                    .padding(horizontal = 8.dp, vertical = 6.dp),
                horizontalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                specialKeyRow.forEach { sk ->
                    val isModActive = (sk.key == "__mod_ctrl" && ctrlActive) || (sk.key == "__mod_alt" && altActive)
                    Button(
                        onClick = { sendKey(sk.key) },
                        shape = RoundedCornerShape(8.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (isModActive) Color(0xFF9D4EDD) else Color(0xFF1E1B3A),
                            contentColor = if (isModActive) Color.White else Color(0xFFB8B0D8)
                        ),
                        modifier = Modifier.height(36.dp)
                    ) {
                        Text(sk.label, fontSize = 11.sp, fontWeight = FontWeight.Bold, fontFamily = FontFamily.Monospace)
                    }
                }
            }

            Box(
                modifier = Modifier
                    .weight(1f)
                    .fillMaxWidth()
                    .padding(horizontal = 12.dp, vertical = 4.dp)
                    .background(Color(0xFF0D0A1F), RoundedCornerShape(12.dp))
                    .padding(12.dp),
                contentAlignment = Alignment.TopStart
            ) {
                Text(
                    text = "$ _",
                    color = Color(0xFF9D4EDD),
                    fontSize = 15.sp,
                    fontFamily = FontFamily.Monospace,
                    fontWeight = FontWeight.Bold
                )
                Text(
                    text = "\n\nToca aquí para escribir",
                    color = Color(0xFF403B60),
                    fontSize = 13.sp,
                    fontFamily = FontFamily.Monospace,
                    modifier = Modifier.align(Alignment.Center)
                )
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .horizontalScroll(rememberScrollState())
                    .padding(horizontal = 8.dp, vertical = 6.dp),
                horizontalArrangement = Arrangement.spacedBy(6.dp)
            ) {
                quickActions.forEach { qa ->
                    Button(
                        onClick = { sendKey(qa.key) },
                        shape = RoundedCornerShape(8.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = Color(0xFF2A1F4A),
                            contentColor = Color(0xFFD8B4FE)
                        ),
                        modifier = Modifier.height(38.dp)
                    ) {
                        Text(qa.label, fontSize = 11.sp, fontWeight = FontWeight.SemiBold, fontFamily = FontFamily.Monospace)
                    }
                }
            }

            Spacer(modifier = Modifier.height(4.dp))
        }

        TextField(
            value = textState,
            onValueChange = { newValue ->
                val newText = newValue.text
                val oldText = textState.text
                when {
                    newText.length < oldText.length -> {
                        val diff = oldText.length - newText.length
                        repeat(diff) {
                            client.sendKey("BackSpace")
                        }
                        if (newText.length < 20) {
                            textState = TextFieldValue(" ".repeat(bufferLimit), TextRange(bufferLimit))
                        } else {
                            textState = newValue
                        }
                    }
                    newText.length > oldText.length -> {
                        val added = newText.substring(oldText.length)
                        when {
                            added == "\n" -> client.sendKey("Return")
                            added == " " -> client.sendKey("space")
                            added.isNotEmpty() -> {
                                if (ctrlActive) {
                                    client.sendKeyCombo("ctrl", added.lowercase())
                                    ctrlActive = false
                                } else if (altActive) {
                                    client.sendKeyCombo("alt", added.lowercase())
                                    altActive = false
                                } else {
                                    client.sendType(added)
                                }
                            }
                        }
                        if (newText.length > 180) {
                            textState = TextFieldValue(" ".repeat(bufferLimit), TextRange(bufferLimit))
                        } else {
                            textState = newValue
                        }
                    }
                    else -> textState = newValue
                }
            },
            modifier = Modifier
                .width(1.dp)
                .height(1.dp)
                .alpha(0f)
                .focusRequester(focusRequester),
            keyboardOptions = KeyboardOptions(
                imeAction = ImeAction.Default,
                capitalization = KeyboardCapitalization.None,
                autoCorrect = false
            )
        )
    }
}
