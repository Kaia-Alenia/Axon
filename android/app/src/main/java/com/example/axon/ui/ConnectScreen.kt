package com.example.axon.ui

import android.Manifest
import android.bluetooth.BluetoothAdapter
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.animation.AnimatedVisibility
import androidx.compose.animation.fadeIn
import androidx.compose.animation.fadeOut
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.navigationBarsPadding
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.statusBarsPadding
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
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
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.core.content.ContextCompat
import com.example.axon.R

private const val PREFS_NAME = "axon_prefs"
private const val KEY_LAST_IP = "last_ip"
private const val KEY_LAST_TOKEN = "last_token"

private fun saveCredentials(context: Context, ip: String, token: String) {
    context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE).edit()
        .putString(KEY_LAST_IP, ip)
        .putString(KEY_LAST_TOKEN, token)
        .apply()
}

private fun loadCredentials(context: Context): Pair<String, String> {
    val prefs = context.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
    return Pair(
        prefs.getString(KEY_LAST_IP, "192.168.1.100") ?: "192.168.1.100",
        prefs.getString(KEY_LAST_TOKEN, "") ?: ""
    )
}

enum class ConnectionMode {
    WiFi, USB, Bluetooth
}

private val BgGradStart  = Color(0xFF0F172A)
private val BgGradEnd    = Color(0xFF020617)
private val AccentGlow   = Color(0xFF1E3A8A)
private val BlueAccent   = Color(0xFF3B82F6)
private val VioletAccent = Color(0xFF8B5CF6)
private val TextPrimary  = Color(0xFFF9FAFB)
private val TextSecond   = Color(0xFF94A3B8)
private val CardBg       = Color(0xFF0F172A)
private val CardInner    = Color(0xFF1E293B)
private val BorderColor  = Color(0xFF334155)
private val ErrorColor   = Color(0xFFEF4444)
private val SuccessColor = Color(0xFF10B981)

@Composable
fun ConnectScreen(
    onConnect: (String) -> Unit,
    onConnectBluetooth: (String) -> Unit,
    onScanQr: () -> Unit,
    onBack: () -> Unit,
    errorMessage: String
) {
    val context = LocalContext.current
    val (savedIp, savedToken) = remember { loadCredentials(context) }

    var mode by remember { mutableStateOf(ConnectionMode.WiFi) }
    var ipInput by remember { mutableStateOf(savedIp) }
    var tokenInput by remember { mutableStateOf(savedToken) }
    var btMac by remember { mutableStateOf("") }

    val bluetoothAdapter = remember { BluetoothAdapter.getDefaultAdapter() }

    var bluetoothPermissionGranted by remember {
        mutableStateOf(
            if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                ContextCompat.checkSelfPermission(context, Manifest.permission.BLUETOOTH_CONNECT) == PackageManager.PERMISSION_GRANTED
            } else {
                true
            }
        )
    }

    val permissionLauncher = rememberLauncherForActivityResult(
        contract = ActivityResultContracts.RequestPermission()
    ) { isGranted ->
        bluetoothPermissionGranted = isGranted
    }

    val pairedDevices by remember(bluetoothPermissionGranted) {
        mutableStateOf(
            try {
                if (bluetoothPermissionGranted) {
                    bluetoothAdapter?.bondedDevices?.toList() ?: emptyList()
                } else {
                    emptyList()
                }
            } catch (e: SecurityException) {
                emptyList<android.bluetooth.BluetoothDevice>()
            }
        )
    }

    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(BgGradStart, BgGradEnd)
                )
            )
            .statusBarsPadding()
            .navigationBarsPadding()
            .padding(horizontal = 24.dp)
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            drawCircle(
                color = AccentGlow.copy(alpha = 0.25f),
                radius = size.minDimension * 0.6f,
                center = Offset(size.width / 2f, size.height * 0.2f)
            )
        }

        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState()),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(24.dp))

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 24.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
              ) {
                Box(
                    modifier = Modifier
                        .size(44.dp)
                        .clip(RoundedCornerShape(14.dp))
                        .background(Color.White.copy(alpha = 0.06f))
                        .border(1.dp, BorderColor, RoundedCornerShape(14.dp))
                        .clickable { onBack() },
                    contentAlignment = Alignment.Center
                ) {
                    Canvas(modifier = Modifier.size(16.dp)) {
                        val w = size.width
                        val h = size.height
                        drawLine(Color.White, start = Offset(w, h / 2), end = Offset(0f, h / 2), strokeWidth = 2.5.dp.toPx())
                        drawLine(Color.White, start = Offset(w * 0.45f, 0f), end = Offset(0f, h / 2), strokeWidth = 2.5.dp.toPx())
                        drawLine(Color.White, start = Offset(w * 0.45f, h), end = Offset(0f, h / 2), strokeWidth = 2.5.dp.toPx())
                    }
                }
                Text(
                    text = stringResource(id = R.string.connection_title),
                    fontSize = 18.sp,
                    fontWeight = FontWeight.Black,
                    color = TextPrimary,
                    letterSpacing = 2.sp
                )
                Spacer(modifier = Modifier.width(44.dp))
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 24.dp)
                    .clip(RoundedCornerShape(16.dp))
                    .background(Color.White.copy(alpha = 0.05f))
                    .border(1.dp, BorderColor, RoundedCornerShape(16.dp))
                    .padding(6.dp),
                horizontalArrangement = Arrangement.SpaceEvenly
            ) {
                listOf(ConnectionMode.WiFi, ConnectionMode.USB, ConnectionMode.Bluetooth).forEach { m ->
                    val selected = mode == m
                    val modeLabel = when (m) {
                        ConnectionMode.WiFi -> stringResource(id = R.string.mode_wifi)
                        ConnectionMode.USB -> stringResource(id = R.string.mode_usb)
                        ConnectionMode.Bluetooth -> stringResource(id = R.string.mode_bluetooth)
                    }
                    Box(
                        modifier = Modifier
                            .weight(1f)
                            .clip(RoundedCornerShape(12.dp))
                            .background(
                                if (selected) {
                                    Brush.horizontalGradient(colors = listOf(BlueAccent, VioletAccent))
                                } else {
                                    Brush.horizontalGradient(colors = listOf(Color.Transparent, Color.Transparent))
                                }
                            )
                            .clickable { mode = m }
                            .padding(vertical = 12.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = modeLabel,
                            color = if (selected) Color.White else TextSecond,
                            fontSize = 13.sp,
                            fontWeight = if (selected) FontWeight.Bold else FontWeight.Medium
                        )
                    }
                }
            }

            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .border(1.dp, BorderColor, RoundedCornerShape(24.dp)),
                shape = RoundedCornerShape(24.dp),
                colors = CardDefaults.cardColors(containerColor = CardBg.copy(alpha = 0.9f))
            ) {
                Column(
                    modifier = Modifier
                        .fillMaxWidth()
                        .padding(24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    when (mode) {
                        ConnectionMode.WiFi -> {
                            Text(
                                text = stringResource(id = R.string.wifi_recommended),
                                color = TextPrimary,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.align(Alignment.Start)
                            )
                            if (tokenInput.isNotEmpty()) {
                                Spacer(modifier = Modifier.height(12.dp))
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(SuccessColor.copy(alpha = 0.1f), RoundedCornerShape(12.dp))
                                        .border(1.dp, SuccessColor.copy(alpha = 0.25f), RoundedCornerShape(12.dp))
                                        .padding(12.dp)
                                ) {
                                    Text(
                                        text = stringResource(id = R.string.saved_token_msg),
                                        color = SuccessColor,
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.SemiBold
                                    )
                                }
                            }
                            Spacer(modifier = Modifier.height(16.dp))
                            OutlinedTextField(
                                value = ipInput,
                                onValueChange = { ipInput = it },
                                label = { Text(stringResource(id = R.string.pc_ip_label), color = TextSecond) },
                                placeholder = { Text(stringResource(id = R.string.pc_ip_placeholder), color = BorderColor) },
                                modifier = Modifier.fillMaxWidth(),
                                singleLine = true,
                                colors = OutlinedTextFieldDefaults.colors(
                                    focusedBorderColor = BlueAccent,
                                    unfocusedBorderColor = BorderColor,
                                    focusedLabelColor = BlueAccent,
                                    unfocusedLabelColor = TextSecond,
                                    focusedTextColor = TextPrimary,
                                    unfocusedTextColor = TextPrimary,
                                    cursorColor = BlueAccent
                                ),
                                shape = RoundedCornerShape(14.dp)
                            )
                            Spacer(modifier = Modifier.height(12.dp))
                            OutlinedTextField(
                                value = tokenInput,
                                onValueChange = { tokenInput = it },
                                label = { Text(stringResource(id = R.string.token_label), color = TextSecond) },
                                placeholder = { Text(stringResource(id = R.string.token_placeholder), color = BorderColor) },
                                modifier = Modifier.fillMaxWidth(),
                                singleLine = true,
                                colors = OutlinedTextFieldDefaults.colors(
                                    focusedBorderColor = BlueAccent,
                                    unfocusedBorderColor = BorderColor,
                                    focusedLabelColor = BlueAccent,
                                    unfocusedLabelColor = TextSecond,
                                    focusedTextColor = TextPrimary,
                                    unfocusedTextColor = TextPrimary,
                                    cursorColor = BlueAccent
                                ),
                                shape = RoundedCornerShape(14.dp)
                            )
                            Spacer(modifier = Modifier.height(8.dp))
                            Text(
                                text = stringResource(id = R.string.token_saved_hint),
                                color = TextSecond,
                                fontSize = 11.sp,
                                modifier = Modifier.align(Alignment.Start),
                                lineHeight = 16.sp
                            )
                            Spacer(modifier = Modifier.height(24.dp))
                            Button(
                                onClick = {
                                    val cleanIp = ipInput.trim()
                                    val cleanToken = tokenInput.trim()
                                    val host = if (cleanIp.contains(":")) cleanIp else "$cleanIp:6969"
                                    val tokenParam = if (cleanToken.isNotEmpty()) "?token=$cleanToken" else ""
                                    val targetUrl = "ws://$host/ws$tokenParam"
                                    saveCredentials(context, cleanIp, cleanToken)
                                    onConnect(targetUrl)
                                },
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(56.dp),
                                shape = RoundedCornerShape(16.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = Color.Transparent,
                                    contentColor = Color.White
                                ),
                                contentPadding = PaddingValues()
                            ) {
                                Box(
                                    modifier = Modifier
                                        .fillMaxSize()
                                        .background(
                                            Brush.horizontalGradient(
                                                colors = listOf(BlueAccent, VioletAccent)
                                            )
                                        ),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Text(stringResource(id = R.string.connect_wifi), fontSize = 15.sp, fontWeight = FontWeight.Bold)
                                }
                            }
                            Spacer(modifier = Modifier.height(12.dp))
                            Button(
                                onClick = onScanQr,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(52.dp)
                                    .border(1.dp, BlueAccent.copy(alpha = 0.6f), RoundedCornerShape(16.dp)),
                                shape = RoundedCornerShape(16.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = Color.White.copy(alpha = 0.05f),
                                    contentColor = TextPrimary
                                )
                            ) {
                                Text(stringResource(id = R.string.scan_qr), fontSize = 15.sp, fontWeight = FontWeight.Bold)
                            }
                        }

                        ConnectionMode.USB -> {
                            Text(
                                text = stringResource(id = R.string.usb_min_latency),
                                color = TextPrimary,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.align(Alignment.Start)
                            )
                            Spacer(modifier = Modifier.height(12.dp))
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(BlueAccent.copy(alpha = 0.1f), RoundedCornerShape(12.dp))
                                    .border(1.dp, BlueAccent.copy(alpha = 0.25f), RoundedCornerShape(12.dp))
                                    .padding(12.dp)
                            ) {
                                Text(
                                    text = stringResource(id = R.string.usb_desc),
                                    color = BlueAccent,
                                    fontSize = 12.sp,
                                    lineHeight = 16.sp
                                )
                            }
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(
                                text = stringResource(id = R.string.usb_instructions),
                                color = TextSecond,
                                fontSize = 13.sp,
                                modifier = Modifier.padding(bottom = 20.dp),
                                lineHeight = 18.sp
                            )
                            Button(
                                onClick = { onConnect("ws://127.0.0.1:6969/ws") },
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(56.dp),
                                shape = RoundedCornerShape(16.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = Color.Transparent,
                                    contentColor = Color.White
                                ),
                                contentPadding = PaddingValues()
                            ) {
                                Box(
                                    modifier = Modifier
                                        .fillMaxSize()
                                        .background(
                                            Brush.horizontalGradient(
                                                colors = listOf(BlueAccent, VioletAccent)
                                            )
                                        ),
                                    contentAlignment = Alignment.Center
                                ) {
                                    Text(stringResource(id = R.string.connect_usb), fontSize = 15.sp, fontWeight = FontWeight.Bold)
                                }
                            }
                        }

                        ConnectionMode.Bluetooth -> {
                            Text(
                                text = stringResource(id = R.string.bt_no_wifi),
                                color = TextPrimary,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.align(Alignment.Start)
                            )
                            Spacer(modifier = Modifier.height(12.dp))
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(TextSecond.copy(alpha = 0.08f), RoundedCornerShape(12.dp))
                                    .border(1.dp, BorderColor, RoundedCornerShape(12.dp))
                                    .padding(12.dp)
                            ) {
                                Text(
                                    text = stringResource(id = R.string.bt_desc),
                                    color = TextSecond,
                                    fontSize = 12.sp,
                                    lineHeight = 16.sp
                                )
                            }
                            Spacer(modifier = Modifier.height(16.dp))

                            if (bluetoothAdapter == null) {
                                Text(
                                    text = stringResource(id = R.string.bt_not_available),
                                    color = ErrorColor,
                                    fontSize = 13.sp,
                                    textAlign = TextAlign.Center
                                )
                            } else if (!bluetoothPermissionGranted && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                                Text(
                                    text = stringResource(id = R.string.bt_permission_required),
                                    color = TextSecond,
                                    fontSize = 13.sp,
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier.padding(bottom = 16.dp)
                                )
                                Button(
                                    onClick = { permissionLauncher.launch(Manifest.permission.BLUETOOTH_CONNECT) },
                                    colors = ButtonDefaults.buttonColors(containerColor = BlueAccent),
                                    shape = RoundedCornerShape(12.dp)
                                ) {
                                    Text(stringResource(id = R.string.grant_permission), color = Color.White, fontWeight = FontWeight.Bold)
                                }
                            } else {
                                Text(
                                    text = stringResource(id = R.string.paired_devices),
                                    color = TextPrimary,
                                    fontSize = 13.sp,
                                    fontWeight = FontWeight.Bold,
                                    modifier = Modifier
                                        .align(Alignment.Start)
                                        .padding(bottom = 10.dp)
                                )
                                if (pairedDevices.isEmpty()) {
                                    Text(
                                        text = stringResource(id = R.string.no_devices_found),
                                        color = TextSecond,
                                        fontSize = 13.sp,
                                        textAlign = TextAlign.Center,
                                        modifier = Modifier.padding(vertical = 16.dp)
                                    )
                                } else {
                                    Column(
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .clip(RoundedCornerShape(16.dp))
                                            .background(CardInner)
                                            .border(1.dp, BorderColor, RoundedCornerShape(16.dp))
                                            .padding(4.dp)
                                    ) {
                                        pairedDevices.forEach { device ->
                                            val deviceName = try { device.name } catch (e: SecurityException) { "Device" }
                                            val deviceAddress = device.address
                                            Row(
                                                modifier = Modifier
                                                    .fillMaxWidth()
                                                    .clickable { onConnectBluetooth(deviceAddress) }
                                                    .padding(horizontal = 16.dp, vertical = 12.dp),
                                                verticalAlignment = Alignment.CenterVertically
                                            ) {
                                                Column(modifier = Modifier.weight(1f)) {
                                                    Text(
                                                        text = deviceName ?: "Unknown",
                                                        color = TextPrimary,
                                                        fontSize = 14.sp,
                                                        fontWeight = FontWeight.SemiBold
                                                    )
                                                    Text(
                                                        text = deviceAddress,
                                                        color = TextSecond,
                                                        fontSize = 11.sp
                                                    )
                                                }
                                                Text(
                                                    text = stringResource(id = R.string.connect_action),
                                                    color = BlueAccent,
                                                    fontSize = 13.sp,
                                                    fontWeight = FontWeight.Bold
                                                )
                                            }
                                        }
                                    }
                                }

                                Spacer(modifier = Modifier.height(16.dp))
                                Text(
                                    text = stringResource(id = R.string.or_enter_mac),
                                    color = TextSecond,
                                    fontSize = 12.sp,
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier.padding(vertical = 6.dp)
                                )
                                OutlinedTextField(
                                    value = btMac,
                                    onValueChange = { btMac = it },
                                    label = { Text(stringResource(id = R.string.bt_mac_label), color = TextSecond) },
                                    placeholder = { Text(stringResource(id = R.string.bt_mac_placeholder), color = BorderColor) },
                                    modifier = Modifier.fillMaxWidth(),
                                    singleLine = true,
                                    colors = OutlinedTextFieldDefaults.colors(
                                        focusedBorderColor = BlueAccent,
                                        unfocusedBorderColor = BorderColor,
                                        focusedLabelColor = BlueAccent,
                                        unfocusedLabelColor = TextSecond,
                                        focusedTextColor = TextPrimary,
                                        unfocusedTextColor = TextPrimary,
                                        cursorColor = BlueAccent
                                    ),
                                    shape = RoundedCornerShape(14.dp)
                                )
                                Spacer(modifier = Modifier.height(16.dp))
                                Button(
                                    onClick = { if (btMac.isNotEmpty()) onConnectBluetooth(btMac.trim()) },
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .height(52.dp),
                                    shape = RoundedCornerShape(16.dp),
                                    colors = ButtonDefaults.buttonColors(
                                        containerColor = Color.Transparent,
                                        contentColor = Color.White
                                    ),
                                    contentPadding = PaddingValues()
                                ) {
                                    Box(
                                        modifier = Modifier
                                            .fillMaxSize()
                                            .background(
                                                Brush.horizontalGradient(
                                                    colors = listOf(BlueAccent, VioletAccent)
                                                )
                                            ),
                                        contentAlignment = Alignment.Center
                                    ) {
                                        Text(stringResource(id = R.string.connect_manually), fontSize = 14.sp, fontWeight = FontWeight.Bold)
                                    }
                                }
                            }
                        }
                    }

                    AnimatedVisibility(
                        visible = errorMessage.isNotEmpty(),
                        enter = fadeIn(),
                        exit = fadeOut()
                    ) {
                        Column {
                            Spacer(modifier = Modifier.height(18.dp))
                            Text(
                                text = errorMessage,
                                color = ErrorColor,
                                fontSize = 13.sp,
                                fontWeight = FontWeight.SemiBold,
                                textAlign = TextAlign.Center,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(ErrorColor.copy(alpha = 0.1f), RoundedCornerShape(12.dp))
                                    .border(1.dp, ErrorColor.copy(alpha = 0.25f), RoundedCornerShape(12.dp))
                                    .padding(12.dp)
                            )
                        }
                    }
                }
            }

            Spacer(modifier = Modifier.height(40.dp))
        }
    }
}
