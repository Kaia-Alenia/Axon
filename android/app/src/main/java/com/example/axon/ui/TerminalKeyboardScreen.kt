package com.example.axon.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
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
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.lazy.LazyColumn
import androidx.compose.foundation.lazy.items
import androidx.compose.foundation.lazy.rememberLazyListState
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.BasicTextField
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.KeyboardType
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateListOf
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalFocusManager
import androidx.compose.ui.platform.LocalSoftwareKeyboardController
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontFamily
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.axon.network.InputClient
import com.example.axon.R

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

private val BgGradStart  = Color(0xFF0F172A)
private val BgGradEnd    = Color(0xFF020617)
private val BlueAccent   = Color(0xFF3B82F6)
private val VioletAccent = Color(0xFF8B5CF6)
private val TextPrimary  = Color(0xFFF9FAFB)
private val TextSecond   = Color(0xFF94A3B8)
private val BorderColor  = Color(0xFF334155)
private val GreenAccent  = Color(0xFF10B981)
private val RedAccent    = Color(0xFFEF4444)

@Composable
fun TerminalKeyboardScreen(client: InputClient, onDisconnect: () -> Unit) {
    val focusRequester = remember { FocusRequester() }
    val focusManager = LocalFocusManager.current
    val keyboardController = LocalSoftwareKeyboardController.current

    val history = remember { mutableStateListOf<String>() }
    var ctrlActive by remember { mutableStateOf(false) }
    var altActive by remember { mutableStateOf(false) }
    val historyListState = rememberLazyListState()
    var textState by remember { mutableStateOf("") }

    BasicTextField(
        value = textState,
        onValueChange = { newText ->
            if (newText.endsWith("\n")) {
                client.sendKey("Return")
                history.add("[Return]")
                textState = ""
            } else if (newText.length > textState.length) {
                val addedText = newText.substring(textState.length)
                client.sendType(addedText)
                history.add(addedText)
                textState = newText
            } else if (newText.length < textState.length) {
                client.sendKey("BackSpace")
                history.add("[BackSpace]")
                textState = newText
            }
        },
        modifier = Modifier
            .size(0.dp)
            .focusRequester(focusRequester),
        keyboardOptions = KeyboardOptions(
            imeAction = ImeAction.Default,
            keyboardType = KeyboardType.Text
        )
    )

    LaunchedEffect(Unit) {
        try { focusRequester.requestFocus(); keyboardController?.show() } catch (_: Exception) {}
    }
    LaunchedEffect(history.size) {
        if (history.isNotEmpty()) historyListState.animateScrollToItem(history.size - 1)
    }

    fun sendKey(key: String) {
        val logKey = when (key) {
            "__mod_ctrl" -> if (!ctrlActive) "[CTRL ON]" else "[CTRL OFF]"
            "__mod_alt"  -> if (!altActive)  "[ALT ON]"  else "[ALT OFF]"
            else -> "key: $key"
        }
        if (key != "__mod_ctrl" && key != "__mod_alt") history.add(logKey)

        when {
            key.startsWith("__combo_") -> {
                val parts = key.removePrefix("__combo_").split("_")
                if (parts.size == 2) client.sendKeyCombo(parts[0], parts[1])
            }
            key == "__mod_ctrl" -> { ctrlActive = !ctrlActive; altActive = false }
            key == "__mod_alt"  -> { altActive = !altActive; ctrlActive = false }
            ctrlActive -> { client.sendKeyCombo("ctrl", key.lowercase()); ctrlActive = false }
            altActive  -> { client.sendKeyCombo("alt", key.lowercase()); altActive = false }
            else -> client.sendKey(key)
        }
    }

    Column(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(BgGradStart, BgGradEnd)
                )
            )
            .padding(12.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(bottom = 10.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                stringResource(id = R.string.terminal_title),
                color = TextPrimary,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
                letterSpacing = (-0.3).sp,
                modifier = Modifier.weight(1f)
            )

            Box(
                modifier = Modifier
                    .clip(RoundedCornerShape(10.dp))
                    .background(RedAccent.copy(alpha = 0.1f))
                    .border(1.dp, RedAccent.copy(alpha = 0.4f), RoundedCornerShape(10.dp))
                    .clickable { onDisconnect() }
                    .padding(horizontal = 12.dp, vertical = 7.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(stringResource(id = R.string.exit_action), color = RedAccent, fontSize = 11.sp, fontWeight = FontWeight.Bold)
            }
        }

        Box(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .clip(RoundedCornerShape(16.dp))
                .background(Color.White.copy(alpha = 0.04f))
                .border(1.dp, BorderColor, RoundedCornerShape(16.dp))
                .clickable {
                    focusManager.clearFocus()
                    focusRequester.requestFocus()
                    keyboardController?.show()
                }
                .padding(14.dp)
        ) {
            LazyColumn(state = historyListState, modifier = Modifier.fillMaxSize()) {
                item {
                    Text(
                        stringResource(id = R.string.terminal_bridge_msg, client.getServerIp()),
                        color = TextSecond.copy(alpha = 0.5f),
                        fontSize = 11.sp,
                        fontFamily = FontFamily.Monospace,
                        lineHeight = 16.sp
                    )
                }
                items(history) { log ->
                    Text(
                        text = log,
                        color = if (log.startsWith("key:") || log.startsWith("[")) TextSecond else GreenAccent,
                        fontSize = 12.sp,
                        fontFamily = FontFamily.Monospace,
                        lineHeight = 18.sp
                    )
                }
                item {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text("axon:~$ ", color = BlueAccent, fontSize = 12.sp, fontFamily = FontFamily.Monospace, fontWeight = FontWeight.Bold)
                        Text("█", color = BlueAccent.copy(alpha = 0.7f), fontSize = 12.sp, fontFamily = FontFamily.Monospace)
                    }
                }
            }
        }

        Spacer(modifier = Modifier.height(10.dp))

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            specialKeyRow.forEach { sk ->
                val isModActive = (sk.key == "__mod_ctrl" && ctrlActive) || (sk.key == "__mod_alt" && altActive)
                Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(10.dp))
                        .background(
                            if (isModActive) {
                                Brush.horizontalGradient(colors = listOf(BlueAccent, VioletAccent))
                            } else {
                                Brush.horizontalGradient(colors = listOf(Color.White.copy(alpha = 0.05f), Color.White.copy(alpha = 0.05f)))
                            }
                        )
                        .border(1.dp, if (isModActive) Color.Transparent else BorderColor, RoundedCornerShape(10.dp))
                        .clickable { sendKey(sk.key) }
                        .padding(horizontal = 14.dp, vertical = 10.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        sk.label,
                        color = if (isModActive) Color.White else TextSecond,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Bold,
                        fontFamily = FontFamily.Monospace
                    )
                }
            }
        }

        Spacer(modifier = Modifier.height(6.dp))

        Row(
            modifier = Modifier
                .fillMaxWidth()
                .horizontalScroll(rememberScrollState()),
            horizontalArrangement = Arrangement.spacedBy(6.dp)
        ) {
            quickActions.forEach { qa ->
                Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(10.dp))
                        .background(BlueAccent.copy(alpha = 0.1f))
                        .border(1.dp, BlueAccent.copy(alpha = 0.25f), RoundedCornerShape(10.dp))
                        .clickable { sendKey(qa.key) }
                        .padding(horizontal = 14.dp, vertical = 10.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        qa.label,
                        color = BlueAccent,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Bold,
                        fontFamily = FontFamily.Monospace
                    )
                }
            }
        }

        Spacer(modifier = Modifier.height(10.dp))

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(12.dp))
                .background(Color.White.copy(alpha = 0.04f))
                .border(1.dp, BorderColor, RoundedCornerShape(12.dp))
                .padding(12.dp),
            contentAlignment = Alignment.Center
        ) {
            Text(
                stringResource(id = R.string.tap_to_activate),
                color = TextSecond,
                fontSize = 11.sp,
                fontFamily = FontFamily.Monospace
            )
        }
    }
}
