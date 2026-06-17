package com.example.aleniaaxon.ui

import androidx.compose.foundation.background
import androidx.compose.foundation.gestures.detectDragGestures
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.text.KeyboardActions
import androidx.compose.foundation.text.KeyboardOptions
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Switch
import androidx.compose.material3.SwitchDefaults
import androidx.compose.material3.Text
import androidx.compose.material3.TextField
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberUpdatedState
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.focus.FocusRequester
import androidx.compose.ui.focus.focusRequester
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.input.pointer.pointerInput
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.text.TextRange
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.input.ImeAction
import androidx.compose.ui.text.input.TextFieldValue
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import android.content.Context

import com.example.aleniaaxon.network.UdpClient
import com.example.aleniaaxon.network.InputClient
import kotlin.math.abs

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

    val accDx = remember { java.util.concurrent.atomic.AtomicLong(java.lang.Double.doubleToRawLongBits(0.0)) }
    val accDy = remember { java.util.concurrent.atomic.AtomicLong(java.lang.Double.doubleToRawLongBits(0.0)) }
    val lastSendTime = remember { java.util.concurrent.atomic.AtomicLong(0L) }
    val throttleMs = 8L

    val sendMove = { dx: Double, dy: Double ->
        if (isLocalConnection) client.sendMove(dx, dy)
        else udpClient.sendMove(dx, dy)
    }

    val sendScroll = { dy: Double ->
        if (isLocalConnection) client.sendScroll(dy)
        else udpClient.sendScroll(dy)
    }

    val currentSendMove by rememberUpdatedState(sendMove)



    val focusRequester = remember { FocusRequester() }
    var textState by remember { mutableStateOf(TextFieldValue("  ", TextRange(2))) }

    val bgBrush = Brush.verticalGradient(
        colors = listOf(Color(0xFF0F0C20), Color(0xFF15102A), Color(0xFF06020F))
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(bgBrush)
            .padding(16.dp)
    ) {
        Column(modifier = Modifier.fillMaxSize()) {
            Row(
                modifier = Modifier.fillMaxWidth().padding(vertical = 8.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                Text(
                    text = "Control Remoto",
                    color = Color.White,
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Bold,
                    modifier = Modifier.weight(1f)
                )
                Button(
                    onClick = onDisconnect,
                    colors = ButtonDefaults.buttonColors(containerColor = Color(0xFFE74C3C)),
                    shape = RoundedCornerShape(10.dp)
                ) {
                    Text("Salir", color = Color.White, fontSize = 12.sp)
                }
            }



            Spacer(modifier = Modifier.height(8.dp))

            Row(modifier = Modifier.weight(1f).fillMaxWidth()) {
                Box(
                    modifier = Modifier
                        .weight(1f)
                        .fillMaxHeight()
                        .background(Color(0xFF1E1B3A), RoundedCornerShape(20.dp))
                        .pointerInput(Unit) {
                            detectTapGestures(onTap = { client.sendClick("left") })
                        }
                        .pointerInput(isLocalConnection) {
                            detectDragGestures(
                                onDragStart = {
                                    accDx.set(java.lang.Double.doubleToRawLongBits(0.0))
                                    accDy.set(java.lang.Double.doubleToRawLongBits(0.0))
                                },
                                onDrag = { change, dragAmount ->
                                    change.consume()
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
                                            val dx = java.lang.Double.longBitsToDouble(accDx.getAndSet(java.lang.Double.doubleToRawLongBits(0.0)))
                                            val dy = java.lang.Double.longBitsToDouble(accDy.getAndSet(java.lang.Double.doubleToRawLongBits(0.0)))
                                            sendMove(dx, dy)
                                        }
                                    }
                                },
                                onDragEnd = {
                                    val dx = java.lang.Double.longBitsToDouble(accDx.getAndSet(java.lang.Double.doubleToRawLongBits(0.0)))
                                    val dy = java.lang.Double.longBitsToDouble(accDy.getAndSet(java.lang.Double.doubleToRawLongBits(0.0)))
                                    if (dx != 0.0 || dy != 0.0) sendMove(dx, dy)
                                }
                            )
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "TOUCHPAD",
                        color = Color(0xFF504B70),
                        fontSize = 14.sp,
                        letterSpacing = 2.sp,
                        fontWeight = FontWeight.Bold
                    )
                }

                Spacer(modifier = Modifier.width(12.dp))

                val scrollAcc = remember { java.util.concurrent.atomic.AtomicReference(0f) }

                Box(
                    modifier = Modifier
                        .width(64.dp)
                        .fillMaxHeight()
                        .background(Color(0xFF15122A), RoundedCornerShape(20.dp))
                        .pointerInput(Unit) {
                            detectDragGestures(
                                onDragStart = { scrollAcc.set(0f) },
                                onDrag = { change, dragAmount ->
                                    change.consume()
                                    val acc = scrollAcc.get() + dragAmount.y
                                    scrollAcc.set(acc)
                                    if (abs(acc) > 12f) {
                                        sendScroll(if (acc > 0) 1.0 else -1.0)
                                        scrollAcc.set(0f)
                                    }
                                }
                            )
                        },
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = "SCROLL",
                        color = Color(0xFF403B60),
                        fontSize = 12.sp,
                        letterSpacing = 1.5.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            Row(modifier = Modifier.fillMaxWidth().height(80.dp)) {
                Button(
                    onClick = { client.sendClick("left") },
                    modifier = Modifier.weight(1f).fillMaxHeight(),
                    shape = RoundedCornerShape(16.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF2E2A4A),
                        contentColor = Color.White
                    )
                ) {
                    Text("Clic Izq", fontSize = 16.sp, fontWeight = FontWeight.Bold)
                }

                Spacer(modifier = Modifier.width(12.dp))

                Button(
                    onClick = { client.sendClick("right") },
                    modifier = Modifier.weight(1f).fillMaxHeight(),
                    shape = RoundedCornerShape(16.dp),
                    colors = ButtonDefaults.buttonColors(
                        containerColor = Color(0xFF2E2A4A),
                        contentColor = Color.White
                    )
                ) {
                    Text("Clic Der", fontSize = 16.sp, fontWeight = FontWeight.Bold)
                }
            }

            Spacer(modifier = Modifier.height(16.dp))

            Button(
                onClick = {
                    textState = TextFieldValue("  ", TextRange(2))
                    try {
                        focusRequester.requestFocus()
                    } catch (e: Exception) {}
                },
                modifier = Modifier.fillMaxWidth().height(56.dp),
                shape = RoundedCornerShape(14.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = Color(0xFF9D4EDD),
                    contentColor = Color.White
                )
            ) {
                Text("Abrir Teclado", fontSize = 16.sp, fontWeight = FontWeight.Bold)
            }
        }

        TextField(
            value = textState,
            onValueChange = { newValue ->
                val newText = newValue.text
                val oldText = textState.text
                when {
                    newText.length < oldText.length -> {
                        client.sendKey("BackSpace")
                        textState = TextFieldValue("  ", TextRange(2))
                    }
                    newText.length > oldText.length -> {
                        val added = newText.substring(oldText.length)
                        when {
                            added == "\n" -> client.sendKey("Return")
                            added == " " -> client.sendKey("space")
                            added.isNotEmpty() -> client.sendType(added)
                        }
                        textState = TextFieldValue("  ", TextRange(2))
                    }
                    else -> textState = newValue
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
