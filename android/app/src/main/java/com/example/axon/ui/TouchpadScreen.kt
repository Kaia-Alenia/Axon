package com.example.axon.ui
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.FastOutSlowInEasing
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.gestures.awaitEachGesture
import androidx.compose.foundation.gestures.awaitFirstDown
import androidx.compose.foundation.gestures.detectTapGestures
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
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
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import com.example.axon.network.UdpClient
import com.example.axon.network.InputClient
import com.example.axon.R
import kotlin.math.abs
import kotlinx.coroutines.channels.BufferOverflow
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.launch
private val BgGradStart  = Color(0xFF0F172A)
private val BgGradEnd    = Color(0xFF020617)
private val BlueAccent   = Color(0xFF3B82F6)
private val VioletAccent = Color(0xFF8B5CF6)
private val TextPrimary  = Color(0xFFF9FAFB)
private val TextSecond   = Color(0xFF94A3B8)
private val BorderColor  = Color(0xFF334155)
private val RedAccent    = Color(0xFFEF4444)
@Composable
fun TouchpadScreen(
    client: InputClient,
    onDisconnect: () -> Unit
) {
    val context = LocalContext.current
    val serverIp = client.getServerIp()
    val isLocalConnection = serverIp == "127.0.0.1" || serverIp == "localhost"
    val isBluetooth = client is com.example.axon.network.BluetoothClient
    val useUdp = !isLocalConnection && !isBluetooth && serverIp.isNotEmpty()
    val udpClient = remember(serverIp, useUdp) {
        if (useUdp) UdpClient(serverIp) else null
    }
    DisposableEffect(udpClient) {
        onDispose { udpClient?.close() }
    }
    val accDx = remember { java.util.concurrent.atomic.AtomicLong(0L) }
    val accDy = remember { java.util.concurrent.atomic.AtomicLong(0L) }
    val lastSendTime = remember { java.util.concurrent.atomic.AtomicLong(0L) }
    val throttleMs = if (isLocalConnection) 12L else 16L
    val lastScrollTime = remember { java.util.concurrent.atomic.AtomicLong(0L) }
    val scrollThrottleMs = 20L
    val sendMove = { dx: Double, dy: Double ->
        if (!useUdp || udpClient == null) client.sendMove(dx, dy) else udpClient.sendMove(dx, dy)
    }
    val sendScroll = { dy: Double ->
        if (!useUdp || udpClient == null) client.sendScroll(dy) else udpClient.sendScroll(dy)
    }
    val dragChannel = remember {
        Channel<Unit>(
            capacity = 1,
            onBufferOverflow = BufferOverflow.DROP_OLDEST
        )
    }
    LaunchedEffect(dragChannel) {
        for (signal in dragChannel) {
            val dx = java.lang.Double.longBitsToDouble(accDx.getAndSet(0L))
            val dy = java.lang.Double.longBitsToDouble(accDy.getAndSet(0L))
            if (dx != 0.0 || dy != 0.0) {
                sendMove(dx, dy)
            }
            kotlinx.coroutines.delay(16)
        }
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
            .background(
                Brush.verticalGradient(
                    colors = listOf(BgGradStart, BgGradEnd)
                )
            )
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
                        text = stringResource(id = R.string.touchpad_title),
                        color = TextPrimary,
                        fontSize = 20.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = (-0.3).sp
                    )
                    Text(
                        text = serverIp,
                        color = BlueAccent,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Medium
                    )
                }
                Box(
                    modifier = Modifier
                        .clip(RoundedCornerShape(12.dp))
                        .background(RedAccent.copy(alpha = 0.1f))
                        .border(1.dp, RedAccent.copy(alpha = 0.4f), RoundedCornerShape(12.dp))
                        .clickable { onDisconnect() }
                        .padding(horizontal = 16.dp, vertical = 8.dp),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = stringResource(id = R.string.disconnect_action),
                        color = RedAccent,
                        fontSize = 12.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
            Row(modifier = Modifier.weight(1f).fillMaxWidth()) {
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxHeight()
                        .clip(RoundedCornerShape(24.dp))
                        .background(Color.White.copy(alpha = 0.04f))
                        .border(1.dp, BorderColor, RoundedCornerShape(24.dp))
                        .pointerInput(isLocalConnection) {
                            var lastTapTime = 0L
                            awaitEachGesture {
                                val down = awaitFirstDown(requireUnconsumed = false)
                                val downTime = System.currentTimeMillis()
                                val isDoubleTap = (downTime - lastTapTime) < 300
                                var isDraggingClick = false
                                touchPos = down.position
                                accDx.set(0L)
                                accDy.set(0L)
                                var dragPointerId = down.id
                                val startTime = System.currentTimeMillis()
                                var isDrag = false
                                var totalDist = 0f
                                var lastZoomDist: Float? = null
                                while (true) {
                                    val event = awaitPointerEvent()
                                    val anyPressed = event.changes.any { it.pressed }
                                    if (!anyPressed) {
                                        touchPos = null
                                        val dx = java.lang.Double.longBitsToDouble(accDx.getAndSet(0L))
                                        val dy = java.lang.Double.longBitsToDouble(accDy.getAndSet(0L))
                                        if (dx != 0.0 || dy != 0.0) sendMove(dx, dy)
                                        if (isDraggingClick) {
                                            client.sendMouseUp("left")
                                        } else {
                                            val duration = System.currentTimeMillis() - startTime
                                            if (!isDrag && duration < 200 && totalDist < 15f) {
                                                client.sendClick("left")
                                                lastTapTime = System.currentTimeMillis()
                                                tapPos = down.position
                                                coroutineScope.launch {
                                                    tapScale.snapTo(0f)
                                                    tapScale.animateTo(
                                                        1f,
                                                        tween(280, easing = FastOutSlowInEasing)
                                                    )
                                                    tapPos = null
                                                }
                                            }
                                        }
                                        break
                                    }
                                    val pressedChanges = event.changes.filter { it.pressed }
                                    if (pressedChanges.size == 2) {
                                        val p1 = pressedChanges[0].position
                                        val p2 = pressedChanges[1].position
                                        val dist = kotlin.math.sqrt((p1.x - p2.x) * (p1.x - p2.x) + (p1.y - p2.y) * (p1.y - p2.y))
                                        if (lastZoomDist != null) {
                                            val delta = dist - lastZoomDist!!
                                            if (abs(delta) > 15f) { 
                                                if (delta > 0) {
                                                    client.sendKeyCombo("ctrl", "+")
                                                } else {
                                                    client.sendKeyCombo("ctrl", "-")
                                                }
                                                lastZoomDist = dist
                                            }
                                        } else {
                                            lastZoomDist = dist
                                        }
                                        event.changes.forEach { it.consume() }
                                    } else {
                                        lastZoomDist = null
                                    }
                                    val change = event.changes.firstOrNull { it.id == dragPointerId && it.pressed }
                                        ?: event.changes.firstOrNull { it.pressed }
                                    if (change != null) {
                                        dragPointerId = change.id
                                        touchPos = change.position
                                        val dragAmount = change.position - change.previousPosition
                                        change.consume()
                                        if (pressedChanges.size == 1) {
                                            val dist = kotlin.math.sqrt(dragAmount.x * dragAmount.x + dragAmount.y * dragAmount.y)
                                            totalDist += dist
                                            if (totalDist > 8f) {
                                                isDrag = true
                                                if (isDoubleTap && !isDraggingClick) {
                                                    isDraggingClick = true
                                                    client.sendMouseDown("left")
                                                }
                                            }
                                            val sensitivity = 2.2
                                            val prevDx = java.lang.Double.longBitsToDouble(accDx.get())
                                            val prevDy = java.lang.Double.longBitsToDouble(accDy.get())
                                            val newDx = prevDx + dragAmount.x.toDouble() * sensitivity
                                            val newDy = prevDy + dragAmount.y.toDouble() * sensitivity
                                            accDx.set(java.lang.Double.doubleToRawLongBits(newDx))
                                            accDy.set(java.lang.Double.doubleToRawLongBits(newDy))
                                            dragChannel.trySend(Unit)
                                        }
                                    }
                                }
                            }
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Canvas(modifier = Modifier.fillMaxSize()) {
                        touchPos?.let { pos ->
                            drawCircle(
                                color = BlueAccent.copy(alpha = 0.15f),
                                radius = 48.dp.toPx(),
                                center = pos
                            )
                            drawCircle(
                                color = BlueAccent.copy(alpha = 0.6f),
                                radius = 5.dp.toPx(),
                                center = pos
                            )
                        }
                        tapPos?.let { pos ->
                            drawCircle(
                                color = BlueAccent.copy(alpha = 0.3f * (1f - tapScale.value)),
                                radius = 38.dp.toPx() * tapScale.value,
                                center = pos,
                                style = Stroke(width = 1.5.dp.toPx())
                            )
                        }
                    }
                    Column(horizontalAlignment = Alignment.CenterHorizontally) {
                        Text(
                            text = "⬡",
                            color = TextSecond.copy(alpha = 0.15f),
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
                        .clip(RoundedCornerShape(24.dp))
                        .background(Color.White.copy(alpha = 0.04f))
                        .border(1.dp, BorderColor, RoundedCornerShape(24.dp))
                        .pointerInput(Unit) {
                            awaitEachGesture {
                                val down = awaitFirstDown(requireUnconsumed = false)
                                scrollAcc.set(0f)
                                var scrollPointerId = down.id
                                while (true) {
                                    val event = awaitPointerEvent()
                                    val anyPressed = event.changes.any { it.pressed }
                                    if (!anyPressed) break
                                    val change = event.changes.firstOrNull { it.id == scrollPointerId && it.pressed }
                                        ?: event.changes.firstOrNull { it.pressed }
                                    if (change != null) {
                                        scrollPointerId = change.id
                                        val dragAmount = change.position - change.previousPosition
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
                                }
                            }
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Canvas(modifier = Modifier.fillMaxSize()) {
                        val step = 22.dp.toPx()
                        val offsetMod = scrollOffset % step
                        var currentY = offsetMod
                        while (currentY < size.height + step) {
                            drawLine(
                                color = BorderColor.copy(alpha = 0.3f),
                                start = Offset(12.dp.toPx(), currentY),
                                end = Offset(size.width - 12.dp.toPx(), currentY),
                                strokeWidth = 1f
                            )
                            currentY += step
                        }
                    }
                    Text(
                        text = stringResource(id = R.string.scroll_label),
                        color = TextSecond.copy(alpha = 0.4f),
                        fontSize = 10.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = 2.sp,
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
                        .background(Color.White.copy(alpha = 0.05f))
                        .border(1.dp, BorderColor, RoundedCornerShape(16.dp))
                        .pointerInput(Unit) {
                            detectTapGestures(
                                onPress = {
                                    client.sendMouseDown("left")
                                    tryAwaitRelease()
                                    client.sendMouseUp("left")
                                }
                            )
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = stringResource(id = R.string.left_click),
                        color = TextPrimary,
                        fontSize = 15.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
                Spacer(modifier = Modifier.width(12.dp))
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxHeight()
                        .clip(RoundedCornerShape(16.dp))
                        .background(Color.White.copy(alpha = 0.05f))
                        .border(1.dp, BorderColor, RoundedCornerShape(16.dp))
                        .pointerInput(Unit) {
                            detectTapGestures(
                                onPress = {
                                    client.sendMouseDown("right")
                                    tryAwaitRelease()
                                    client.sendMouseUp("right")
                                }
                            )
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = stringResource(id = R.string.right_click),
                        color = TextPrimary,
                        fontSize = 15.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
        TextField(
            value = textState,
            onValueChange = { newValue ->
                val newText = newValue.text
                val oldText = textState.text
                textState = TextFieldValue(" ", TextRange(1))
                when {
                    newText.length < oldText.length -> client.sendKey("BackSpace")
                    newText.length > oldText.length -> {
                        val added = if (newText.startsWith(oldText)) newText.substring(oldText.length) else newText
                        added.forEach { ch ->
                            when {
                                ch == '\n' -> client.sendKey("Return")
                                ch == ' '  -> client.sendKey("space")
                                else       -> client.sendType(ch.toString())
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
            keyboardOptions = KeyboardOptions(
                imeAction = ImeAction.Default,
                autoCorrectEnabled = false,
                keyboardType = androidx.compose.ui.text.input.KeyboardType.Password
            ),
            keyboardActions = KeyboardActions(onAny = {
                client.sendKey("Return")
                textState = TextFieldValue(" ", TextRange(1))
            })
        )
    }
}
