package com.example.axon.ui

import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.detectDragGestures
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.draw.rotate
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.axon.network.UdpClient
import com.example.axon.network.InputClient
import kotlin.math.abs
import kotlinx.coroutines.launch

private val BgColor2     = Color(0xFF0B1120)
private val CardBg2      = Color(0xFF111827)
private val Blue5002     = Color(0xFF3B82F6)
private val Blue3002     = Color(0xFF93C5FD)
private val TextPrim2    = Color(0xFFF9FAFB)
private val TextSec2     = Color(0xFF9CA3AF)
private val Border2      = Color(0xFF1F2937)
private val RedBtn       = Color(0xFF7F1D1D)
private val RedBorder    = Color(0xFFEF4444)

@Composable
fun TouchpadScreen(
    client: InputClient,
    onDisconnect: () -> Unit
) {
    val context = LocalContext.current
    val serverIp = client.getServerIp()
    val udpClient = remember(serverIp) { UdpClient(serverIp) }

    DisposableEffect(udpClient) {
        onDispose { udpClient.close() }
    }

    val isLocalConnection = serverIp == "127.0.0.1" || serverIp == "localhost"
    val isBluetooth = client is com.example.axon.network.BluetoothClient
    val useUdp = !isLocalConnection && !isBluetooth

    val accDx = remember { java.util.concurrent.atomic.AtomicLong(java.lang.Double.doubleToRawLongBits(0.0)) }
    val accDy = remember { java.util.concurrent.atomic.AtomicLong(java.lang.Double.doubleToRawLongBits(0.0)) }
    val lastSendTime = remember { java.util.concurrent.atomic.AtomicLong(0L) }
    val throttleMs = if (isLocalConnection) 12L else 16L
    val lastScrollTime = remember { java.util.concurrent.atomic.AtomicLong(0L) }
    val scrollThrottleMs = 20L

    val sendMove = { dx: Double, dy: Double ->
        if (!useUdp) client.sendMove(dx, dy) else udpClient.sendMove(dx, dy)
    }
    val sendScroll = { dy: Double ->
        if (!useUdp) client.sendScroll(dy) else udpClient.sendScroll(dy)
    }

    val focusRequester = remember { FocusRequester() }
    var textState by remember { mutableStateOf(TextFieldValue("  ", TextRange(2))) }
    val coroutineScope = rememberCoroutineScope()

    var touchPos by remember { mutableStateOf<Offset?>(null) }
    val tapScale = remember { Animatable(0f) }
    var tapPos by remember { mutableStateOf<Offset?>(null) }
    var scrollOffset by remember { mutableStateOf(0f) }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(BgColor2)
            .padding(horizontal = 16.dp, vertical = 12.dp)
    ) {
        Column(modifier = Modifier.fillMaxSize()) {

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 12.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Column(modifier = Modifier.weight(1f)) {
                    Text(
                        text = "Touchpad",
                        color = TextPrim2,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = (-0.3).sp
                    )
                    Text(
                        text = serverIp,
                        color = Blue3002,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Normal
                    )
                }

                Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(10.dp))
                        .background(RedBtn.copy(alpha = 0.3f))
                        .border(1.dp, RedBorder.copy(alpha = 0.5f), RoundedCornerShape(10.dp))
                        .clickable { onDisconnect() }
                        .padding(horizontal = 16.dp, vertical = 8.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "Desconectar",
                        color = RedBorder,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }

            Row(modifier = Modifier.weight(1f).fillMaxWidth()) {

                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxHeight()
                        .clip(RoundedCornerShape(20.dp))
                        .background(CardBg2)
                        .border(1.dp, Border2, RoundedCornerShape(20.dp))
                        .pointerInput(Unit) {
                            detectTapGestures(
                                onTap = { offset ->
                                    client.sendClick("left")
                                    tapPos = offset
                                    coroutineScope.launch {
                                        tapScale.snapTo(0f)
                                        tapScale.animateTo(
                                            1f,
                                            tween(280, easing = FastOutSlowInEasing)
                                        )
                                        tapPos = null
                                    }
                                }
                            )
                        }
                        .pointerInput(isLocalConnection) {
                            detectDragGestures(
                                onDragStart = { offset ->
                                    touchPos = offset
                                    accDx.set(java.lang.Double.doubleToRawLongBits(0.0))
                                    accDy.set(java.lang.Double.doubleToRawLongBits(0.0))
                                },
                                onDrag = { change, dragAmount ->
                                    change.consume()
                                    touchPos = change.position
                                    val sensitivity = 2.2
                                    val prevDx = java.lang.Double.longBitsToDouble(accDx.get())
                                    val prevDy = java.lang.Double.longBitsToDouble(accDy.get())
                                    val newDx = prevDx + dragAmount.x.toDouble() * sensitivity
                                    val newDy = prevDy + dragAmount.y.toDouble() * sensitivity
                                    accDx.set(java.lang.Double.doubleToRawLongBits(newDx))
                                    accDy.set(java.lang.Double.doubleToRawLongBits(newDy))
                                    val now = System.currentTimeMillis()
                                    val last = lastSendTime.get()
                                    if (now - last >= throttleMs) {
                                        if (lastSendTime.compareAndSet(last, now)) {
                                            val dx = java.lang.Double.longBitsToDouble(
                                                accDx.getAndSet(java.lang.Double.doubleToRawLongBits(0.0))
                                            )
                                            val dy = java.lang.Double.longBitsToDouble(
                                                accDy.getAndSet(java.lang.Double.doubleToRawLongBits(0.0))
                                            )
                                            sendMove(dx, dy)
                                        }
                                    }
                                },
                                onDragEnd = {
                                    touchPos = null
                                    val dx = java.lang.Double.longBitsToDouble(
                                        accDx.getAndSet(java.lang.Double.doubleToRawLongBits(0.0))
                                    )
                                    val dy = java.lang.Double.longBitsToDouble(
                                        accDy.getAndSet(java.lang.Double.doubleToRawLongBits(0.0))
                                    )
                                    if (dx != 0.0 || dy != 0.0) sendMove(dx, dy)
                                }
                            )
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Canvas(modifier = Modifier.fillMaxSize()) {
                        touchPos?.let { pos ->
                            drawCircle(
                                color = Blue5002.copy(alpha = 0.15f),
                                radius = 48.dp.toPx(),
                                center = pos
                            )
                            drawCircle(
                                color = Blue5002.copy(alpha = 0.6f),
                                radius = 5.dp.toPx(),
                                center = pos
                            )
                        }
                        tapPos?.let { pos ->
                            drawCircle(
                                color = Blue5002.copy(alpha = 0.3f * (1f - tapScale.value)),
                                radius = 38.dp.toPx() * tapScale.value,
                                center = pos,
                                style = Stroke(width = 1.5.dp.toPx())
                            )
                        }
                    }

                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = "⬡",
                            color = TextSec2.copy(alpha = 0.2f),
                            fontSize = 32.sp
                        )
                    }
                }

                Spacer(modifier = Modifier.width(12.dp))

                val scrollAcc = remember { java.util.concurrent.atomic.AtomicReference(0f) }

                Box(
                    modifier = Modifier
                        .width(62.dp)
                        .fillMaxHeight()
                        .clip(RoundedCornerShape(20.dp))
                        .background(CardBg2)
                        .border(1.dp, Border2, RoundedCornerShape(20.dp))
                        .pointerInput(Unit) {
                            detectDragGestures(
                                onDragStart = { scrollAcc.set(0f) },
                                onDrag = { change, dragAmount ->
                                    change.consume()
                                    scrollOffset += dragAmount.y
                                    val acc = scrollAcc.get() + dragAmount.y
                                    scrollAcc.set(acc)
                                    if (abs(acc) > 18f) {
                                        val now = System.currentTimeMillis()
                                        val last = lastScrollTime.get()
                                        if (now - last >= scrollThrottleMs) {
                                            if (lastScrollTime.compareAndSet(last, now)) {
                                                sendScroll(if (acc > 0) 1.0 else -1.0)
                                                scrollAcc.set(0f)
                                            }
                                        }
                                    }
                                }
                            )
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Canvas(modifier = Modifier.fillMaxSize()) {
                        val step = 22.dp.toPx()
                        val offsetMod = scrollOffset % step
                        var currentY = offsetMod
                        while (currentY < size.height + step) {
                            drawLine(
                                color = Border2.copy(alpha = 0.5f),
                                start = Offset(10.dp.toPx(), currentY),
                                end = Offset(size.width - 10.dp.toPx(), currentY),
                                strokeWidth = 1f
                            )
                            currentY += step
                        }
                    }
                    Text(
                        "SCROLL",
                        color = TextSec2.copy(alpha = 0.4f),
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Medium,
                        letterSpacing = 1.5.sp,
                        modifier = Modifier.rotate(90f)
                    )
                }
            }

            Spacer(modifier = Modifier.height(12.dp))

            Row(modifier = Modifier.fillMaxWidth().height(72.dp)) {
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxHeight()
                        .clip(RoundedCornerShape(16.dp))
                        .background(CardBg2)
                        .border(1.dp, Border2, RoundedCornerShape(16.dp))
                        .clickable { client.sendClick("left") },
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        "Clic Izq",
                        color = TextPrim2,
                        fontSize = 15.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }

                Spacer(modifier = Modifier.width(12.dp))

                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxHeight()
                        .clip(RoundedCornerShape(16.dp))
                        .background(CardBg2)
                        .border(1.dp, Border2, RoundedCornerShape(16.dp))
                        .clickable { client.sendClick("right") },
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        "Clic Der",
                        color = TextPrim2,
                        fontSize = 15.sp,
                        fontWeight = FontWeight.SemiBold
                    )
                }
            }

        }


        TextField(
            value = textState,
            onValueChange = { newValue ->
                val newText = newValue.text
                val oldText = textState.text
                textState = TextFieldValue("  ", TextRange(2))
                when {
                    newText.length < oldText.length -> client.sendKey("BackSpace")
                    newText.length > oldText.length -> {
                        val added = newText.removePrefix(oldText)
                        when {
                            added == "\n" -> client.sendKey("Return")
                            added == " "  -> client.sendKey("space")
                            added.length == 1 -> client.sendType(added)
                            else -> added.forEach { ch ->
                                when {
                                    ch == '\n' -> client.sendKey("Return")
                                    ch == ' '  -> client.sendKey("space")
                                    else       -> client.sendType(ch.toString())
                                }
                            }
                        }
                    }
                }
            },
            modifier = Modifier
                .width(1.dp)
                .height(1.dp)
                .alpha(0f)
                .focusRequester(focusRequester),
            keyboardOptions = KeyboardOptions(imeAction = ImeAction.Default),
            keyboardActions = KeyboardActions(onAny = {
                client.sendKey("Return")
                textState = TextFieldValue("  ", TextRange(2))
            })
        )
    }
}
