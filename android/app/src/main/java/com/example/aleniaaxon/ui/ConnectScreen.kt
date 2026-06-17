package com.example.aleniaaxon.ui

import androidx.compose.foundation.background
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
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.OutlinedTextField
import androidx.compose.material3.OutlinedTextFieldDefaults
import androidx.compose.material3.Text
import androidx.compose.runtime.Composable
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.setValue
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp

enum class ConnectionMode {
    WiFi, USB, Bluetooth
}

@Composable
fun ConnectScreen(
    onConnect: (String) -> Unit,
    onConnectBluetooth: (String) -> Unit,
    onScanQr: () -> Unit,
    errorMessage: String
) {
    var mode by remember { mutableStateOf(ConnectionMode.WiFi) }
    var ipInput by remember { mutableStateOf("192.168.1.100") }
    var tokenInput by remember { mutableStateOf("") }
    
    var btMac by remember { mutableStateOf("") }

    val bgBrush = Brush.verticalGradient(
        colors = listOf(Color(0xFF0F0C20), Color(0xFF15102A), Color(0xFF06020F))
    )

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(bgBrush)
            .padding(24.dp),
        contentAlignment = Alignment.Center
    ) {
        Column(
            modifier = Modifier.fillMaxWidth(),
            horizontalAlignment = Alignment.CenterHorizontally,
            verticalArrangement = Arrangement.Center
        ) {
            Text(
                text = "Alenia AXON",
                fontSize = 32.sp,
                fontWeight = FontWeight.Bold,
                color = Color.White,
                textAlign = TextAlign.Center
            )
            
            Spacer(modifier = Modifier.height(4.dp))
            
            Text(
                text = "Universal Input Bridge",
                fontSize = 14.sp,
                color = Color(0xFFA099C0),
                textAlign = TextAlign.Center
            )

            Spacer(modifier = Modifier.height(28.dp))

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .background(Color(0xFF15122A), RoundedCornerShape(10.dp))
                    .padding(4.dp),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                listOf(ConnectionMode.WiFi, ConnectionMode.USB, ConnectionMode.Bluetooth).forEach { m ->
                    val selected = mode == m
                    Button(
                        onClick = { mode = m },
                        colors = ButtonDefaults.buttonColors(
                            containerColor = if (selected) Color(0xFF9D4EDD) else Color.Transparent,
                            contentColor = if (selected) Color.White else Color(0xFFA099C0)
                        ),
                        shape = RoundedCornerShape(8.dp),
                        modifier = Modifier.weight(1f),
                        contentPadding = PaddingValues(vertical = 8.dp)
                    ) {
                        Text(m.name, fontSize = 13.sp, fontWeight = FontWeight.Bold)
                    }
                }
            }

            Spacer(modifier = Modifier.height(24.dp))

            when (mode) {
                ConnectionMode.WiFi -> {
                    OutlinedTextField(
                        value = ipInput,
                        onValueChange = { ipInput = it },
                        label = { Text("IP de la PC", color = Color(0xFFA099C0)) },
                        placeholder = { Text("Ej. 192.168.1.15", color = Color(0xFF504B70)) },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true,
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = Color(0xFF9D4EDD),
                            unfocusedBorderColor = Color(0xFF302B50),
                            focusedLabelColor = Color(0xFF9D4EDD),
                            unfocusedLabelColor = Color(0xFFA099C0),
                            focusedTextColor = Color.White,
                            unfocusedTextColor = Color.White
                        ),
                        shape = RoundedCornerShape(14.dp)
                    )

                    Spacer(modifier = Modifier.height(16.dp))

                    Button(
                        onClick = {
                            val cleanIp = ipInput.trim()
                            val targetUrl = if (cleanIp.startsWith("ws://") || cleanIp.startsWith("http://")) {
                                var url = cleanIp.replace("http://", "ws://")
                                if (!url.endsWith("/ws")) {
                                    url += if (url.endsWith("/")) "ws" else "/ws"
                                }
                                url
                            } else {
                                val host = if (cleanIp.contains(":")) cleanIp else "$cleanIp:6969"
                                "ws://$host/ws"
                            }
                            onConnect(targetUrl)
                        },
                        modifier = Modifier.fillMaxWidth().height(50.dp),
                        shape = RoundedCornerShape(14.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = Color(0xFF9D4EDD),
                            contentColor = Color.White
                        )
                    ) {
                        Text("Conectar por Wi-Fi", fontSize = 15.sp, fontWeight = FontWeight.Bold)
                    }

                    Spacer(modifier = Modifier.height(12.dp))

                    Button(
                        onClick = onScanQr,
                        modifier = Modifier.fillMaxWidth().height(50.dp),
                        shape = RoundedCornerShape(14.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = Color(0xFF1E1B3A),
                            contentColor = Color(0xFFD8B4FE)
                        )
                    ) {
                        Text("Escanear Código QR", fontSize = 15.sp, fontWeight = FontWeight.Bold)
                    }
                }
                ConnectionMode.USB -> {
                    Text(
                        text = "Conecta el cable USB y activa la depuración USB en Ajustes de Desarrollador.",
                        color = Color(0xFFA099C0),
                        fontSize = 13.sp,
                        textAlign = TextAlign.Center,
                        modifier = Modifier.padding(horizontal = 8.dp)
                    )

                    Spacer(modifier = Modifier.height(16.dp))

                    OutlinedTextField(
                        value = tokenInput,
                        onValueChange = { tokenInput = it },
                        label = { Text("Token de Seguridad", color = Color(0xFFA099C0)) },
                        placeholder = { Text("Token de 8 caracteres", color = Color(0xFF504B70)) },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true,
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = Color(0xFF9D4EDD),
                            unfocusedBorderColor = Color(0xFF302B50),
                            focusedLabelColor = Color(0xFF9D4EDD),
                            unfocusedLabelColor = Color(0xFFA099C0),
                            focusedTextColor = Color.White,
                            unfocusedTextColor = Color.White
                        ),
                        shape = RoundedCornerShape(14.dp)
                    )

                    Spacer(modifier = Modifier.height(16.dp))

                    Button(
                        onClick = {
                            val cleanToken = tokenInput.trim()
                            onConnect("ws://127.0.0.1:6969/ws?token=$cleanToken")
                        },
                        modifier = Modifier.fillMaxWidth().height(50.dp),
                        shape = RoundedCornerShape(14.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = Color(0xFF9D4EDD),
                            contentColor = Color.White
                        )
                    ) {
                        Text("Conectar por USB (ADB)", fontSize = 15.sp, fontWeight = FontWeight.Bold)
                    }
                }
                ConnectionMode.Bluetooth -> {
                    OutlinedTextField(
                        value = btMac,
                        onValueChange = { btMac = it },
                        label = { Text("MAC Address Bluetooth", color = Color(0xFFA099C0)) },
                        placeholder = { Text("Ej. 00:11:22:33:44:55", color = Color(0xFF504B70)) },
                        modifier = Modifier.fillMaxWidth(),
                        singleLine = true,
                        colors = OutlinedTextFieldDefaults.colors(
                            focusedBorderColor = Color(0xFF9D4EDD),
                            unfocusedBorderColor = Color(0xFF302B50),
                            focusedLabelColor = Color(0xFF9D4EDD),
                            unfocusedLabelColor = Color(0xFFA099C0),
                            focusedTextColor = Color.White,
                            unfocusedTextColor = Color.White
                        ),
                        shape = RoundedCornerShape(12.dp)
                    )

                    Spacer(modifier = Modifier.height(16.dp))

                    Button(
                        onClick = {
                            onConnectBluetooth(btMac.trim())
                        },
                        modifier = Modifier.fillMaxWidth().height(50.dp),
                        shape = RoundedCornerShape(14.dp),
                        colors = ButtonDefaults.buttonColors(
                            containerColor = Color(0xFF9D4EDD),
                            contentColor = Color.White
                        )
                    ) {
                        Text("Conectar por Bluetooth", fontSize = 15.sp, fontWeight = FontWeight.Bold)
                    }
                }
            }

            if (errorMessage.isNotEmpty()) {
                Spacer(modifier = Modifier.height(20.dp))
                Text(
                    text = errorMessage,
                    color = Color(0xFFE74C3C),
                    fontSize = 14.sp,
                    textAlign = TextAlign.Center
                )
            }
        }
    }
}
