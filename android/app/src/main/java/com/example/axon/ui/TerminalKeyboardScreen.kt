package com.example.axon.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.horizontalScroll
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
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

private val KbBg       = Color(0xFF0B1120)
private val KbCard     = Color(0xFF111827)
private val KbBlue     = Color(0xFF3B82F6)
private val KbBlueDim  = Color(0xFF1D4ED8)
private val KbTextW    = Color(0xFFF9FAFB)
private val KbTextG    = Color(0xFF9CA3AF)
private val KbBorder   = Color(0xFF1F2937)
private val KbGreen    = Color(0xFF34D399)
private val KbRed      = Color(0xFFEF4444)
private val KbRedBg    = Color(0xFF7F1D1D)

@Composable
fun TerminalKeyboardScreen(client: InputClient, onDisconnect: () -> Unit) {
    val focusRequester = remember { FocusRequester() }
    val focusManager = LocalFocusManager.current
    val keyboardController = LocalSoftwareKeyboardController.current

    val history = remember { mutableStateListOf<String>() }
    var ctrlActive by remember { mutableStateOf(false) }
    var altActive by remember { mutableStateOf(false) }
    val historyListState = rememberLazyListState()

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
            .background(KbBg)
            .padding(12.dp)
    ) {
        Row(
            modifier = Modifier.fillMaxWidth().padding(bottom = 10.dp),
            verticalAlignment = Alignment.CenterVertically
        ) {
            Text(
                stringResource(id = R.string.terminal_title),
                color = KbTextW,
                fontSize = 18.sp,
                fontWeight = FontWeight.Bold,
                letterSpacing = (-0.3).sp,
                modifier = Modifier.weight(1f)
            )

            Box(
                modifier = Modifier
                    .clip(RoundedCornerShape(8.dp))
                    .background(KbRedBg.copy(alpha = 0.3f))
                    .border(1.dp, KbRed.copy(alpha = 0.5f), RoundedCornerShape(8.dp))
                    .clickable { onDisconnect() }
                    .padding(horizontal = 12.dp, vertical = 7.dp),
                contentAlignment = Alignment.Center
            ) {
                Text(stringResource(id = R.string.exit_action), color = KbRed, fontSize = 11.sp, fontWeight = FontWeight.SemiBold)
            }
        }

        Box(
            modifier = Modifier
                .weight(1f)
                .fillMaxWidth()
                .clip(RoundedCornerShape(14.dp))
                .background(KbCard)
                .border(1.dp, KbBorder, RoundedCornerShape(14.dp))
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
                        color = KbTextG.copy(alpha = 0.5f),
                        fontSize = 11.sp,
                        fontFamily = FontFamily.Monospace,
                        lineHeight = 16.sp
                    )
                }
                items(history) { log ->
                    Text(
                        text = log,
                        color = if (log.startsWith("key:") || log.startsWith("[")) KbTextG else KbGreen,
                        fontSize = 12.sp,
                        fontFamily = FontFamily.Monospace,
                        lineHeight = 18.sp
                    )
                }
                item {
                    Row(verticalAlignment = Alignment.CenterVertically) {
                        Text("axon:~$ ", color = KbBlue, fontSize = 12.sp, fontFamily = FontFamily.Monospace, fontWeight = FontWeight.Bold)
                        Text("█", color = KbBlue.copy(alpha = 0.7f), fontSize = 12.sp, fontFamily = FontFamily.Monospace)
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
                        .clip(RoundedCornerShape(8.dp))
                        .background(if (isModActive) KbBlue else KbCard)
                        .border(1.dp, if (isModActive) KbBlue else KbBorder, RoundedCornerShape(8.dp))
                        .clickable { sendKey(sk.key) }
                        .padding(horizontal = 13.dp, vertical = 9.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        sk.label,
                        color = if (isModActive) Color.White else KbTextG,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Medium,
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
                        .clip(RoundedCornerShape(8.dp))
                        .background(KbBlueDim.copy(alpha = 0.15f))
                        .border(1.dp, KbBlue.copy(alpha = 0.25f), RoundedCornerShape(8.dp))
                        .clickable { sendKey(qa.key) }
                        .padding(horizontal = 13.dp, vertical = 9.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        qa.label,
                        color = KbBlue,
                        fontSize = 11.sp,
                        fontWeight = FontWeight.Medium,
                        fontFamily = FontFamily.Monospace
                    )
                }
            }
        }

        Spacer(modifier = Modifier.height(8.dp))

        Box(
            modifier = Modifier
                .fillMaxWidth()
                .clip(RoundedCornerShape(10.dp))
                .background(KbCard)
                .border(1.dp, KbBorder, RoundedCornerShape(10.dp))
                .padding(10.dp),
            contentAlignment = Alignment.Center
        ) {
            Text(
                stringResource(id = R.string.tap_to_activate),
                color = KbTextG,
                fontSize = 11.sp,
                fontFamily = FontFamily.Monospace
            )
        }
    }
}

