package com.example.axon.ui

import android.Manifest
import android.bluetooth.BluetoothAdapter
import android.content.Context
import android.content.pm.PackageManager
import android.os.Build
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
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

private val CBg = Color(0xFF0B1120)
private val CPrimary = Color(0xFF3B82F6)
private val CTextWhite = Color(0xFFF9FAFB)
private val CTextSlate = Color(0xFF9CA3AF)
private val CCardBg = Color(0xFF111827)
private val CBorder = Color(0xFF1F2937)
private val CBorderActive = Color(0xFF3B82F6)
private val CError = Color(0xFFEF4444)
private val CSuccess = Color(0xFF34D399)

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
            .background(CBg)
            .statusBarsPadding()
            .navigationBarsPadding()
            .padding(horizontal = 24.dp)
    ) {
        Column(
            modifier = Modifier
                .fillMaxSize()
                .verticalScroll(rememberScrollState()),
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(28.dp))

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 20.dp),
                verticalAlignment = Alignment.CenterVertically,
                horizontalArrangement = Arrangement.SpaceBetween
            ) {
                Box(
                    modifier = Modifier
                        .size(40.dp)
                        .clip(RoundedCornerShape(10.dp))
                        .background(CCardBg)
                        .border(1.dp, CBorder, RoundedCornerShape(10.dp))
                        .clickable { onBack() },
                    contentAlignment = Alignment.Center
                ) {
                    Canvas(modifier = Modifier.size(16.dp)) {
                        val w = size.width
                        val h = size.height
                        drawLine(Color.White, start = Offset(w, h / 2), end = Offset(0f, h / 2), strokeWidth = 2.dp.toPx())
                        drawLine(Color.White, start = Offset(w * 0.4f, 0f), end = Offset(0f, h / 2), strokeWidth = 2.dp.toPx())
                        drawLine(Color.White, start = Offset(w * 0.4f, h), end = Offset(0f, h / 2), strokeWidth = 2.dp.toPx())
                    }
                }
                Text(
                    text = stringResource(id = R.string.connection_title),
                    fontSize = 18.sp,
                    fontWeight = FontWeight.ExtraBold,
                    color = CTextWhite,
                    letterSpacing = 2.sp
                )
                Spacer(modifier = Modifier.width(40.dp))
            }

            Row(
                modifier = Modifier
                    .fillMaxWidth()
                    .padding(bottom = 20.dp)
                    .clip(RoundedCornerShape(12.dp))
                    .background(CCardBg)
                    .border(1.dp, CBorder, RoundedCornerShape(12.dp))
                    .padding(4.dp),
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
                            .clip(RoundedCornerShape(8.dp))
                            .background(if (selected) CPrimary else Color.Transparent)
                            .clickable { mode = m }
                            .padding(vertical = 10.dp),
                        contentAlignment = Alignment.Center
                    ) {
                        Text(
                            text = modeLabel,
                            color = if (selected) Color.White else CTextSlate,
                            fontSize = 13.sp,
                            fontWeight = if (selected) FontWeight.Bold else FontWeight.Medium
                        )
                    }
                }
            }

            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .border(1.dp, CBorder, RoundedCornerShape(16.dp)),
                shape = RoundedCornerShape(16.dp),
                colors = CardDefaults.cardColors(containerColor = CCardBg)
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
                                color = CTextWhite,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.align(Alignment.Start)
                            )
                            if (tokenInput.isNotEmpty()) {
                                Spacer(modifier = Modifier.height(10.dp))
                                Box(
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .background(Color(0x1A238636), RoundedCornerShape(8.dp))
                                        .border(1.dp, Color(0x33238636), RoundedCornerShape(8.dp))
                                        .padding(10.dp)
                                ) {
                                    Text(
                                        text = stringResource(id = R.string.saved_token_msg),
                                        color = Color(0xFF3FB950),
                                        fontSize = 11.sp,
                                        fontWeight = FontWeight.Medium
                                    )
                                }
                            }
                            Spacer(modifier = Modifier.height(16.dp))
                            OutlinedTextField(
                                value = ipInput,
                                onValueChange = { ipInput = it },
                                label = { Text(stringResource(id = R.string.pc_ip_label), color = CTextSlate) },
                                placeholder = { Text(stringResource(id = R.string.pc_ip_placeholder), color = Color(0xFF484F58)) },
                                modifier = Modifier.fillMaxWidth(),
                                singleLine = true,
                                colors = OutlinedTextFieldDefaults.colors(
                                    focusedBorderColor = CPrimary,
                                    unfocusedBorderColor = CBorderActive,
                                    focusedLabelColor = CPrimary,
                                    unfocusedLabelColor = CTextSlate,
                                    focusedTextColor = CTextWhite,
                                    unfocusedTextColor = CTextWhite,
                                    cursorColor = CPrimary
                                ),
                                shape = RoundedCornerShape(12.dp)
                            )
                            Spacer(modifier = Modifier.height(10.dp))
                            OutlinedTextField(
                                value = tokenInput,
                                onValueChange = { tokenInput = it },
                                label = { Text(stringResource(id = R.string.token_label), color = CTextSlate) },
                                placeholder = { Text(stringResource(id = R.string.token_placeholder), color = Color(0xFF484F58)) },
                                modifier = Modifier.fillMaxWidth(),
                                singleLine = true,
                                colors = OutlinedTextFieldDefaults.colors(
                                    focusedBorderColor = CPrimary,
                                    unfocusedBorderColor = CBorderActive,
                                    focusedLabelColor = CPrimary,
                                    unfocusedLabelColor = CTextSlate,
                                    focusedTextColor = CTextWhite,
                                    unfocusedTextColor = CTextWhite,
                                    cursorColor = CPrimary
                                ),
                                shape = RoundedCornerShape(12.dp)
                            )
                            Spacer(modifier = Modifier.height(6.dp))
                            Text(
                                text = stringResource(id = R.string.token_saved_hint),
                                color = CTextSlate,
                                fontSize = 10.sp,
                                modifier = Modifier.align(Alignment.Start)
                            )
                            Spacer(modifier = Modifier.height(20.dp))
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
                                    .height(54.dp),
                                shape = RoundedCornerShape(12.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = CPrimary,
                                    contentColor = Color.White
                                )
                            ) {
                                Text(stringResource(id = R.string.connect_wifi), fontSize = 15.sp, fontWeight = FontWeight.Bold)
                            }
                            Spacer(modifier = Modifier.height(10.dp))
                            Button(
                                onClick = onScanQr,
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(50.dp)
                                    .border(1.dp, CBorderActive, RoundedCornerShape(12.dp)),
                                shape = RoundedCornerShape(12.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = CCardBg,
                                    contentColor = CTextWhite
                                )
                            ) {
                                Text(stringResource(id = R.string.scan_qr), fontSize = 15.sp, fontWeight = FontWeight.SemiBold)
                            }
                        }

                        ConnectionMode.USB -> {
                            Text(
                                text = stringResource(id = R.string.usb_min_latency),
                                color = CTextWhite,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.align(Alignment.Start)
                            )
                            Spacer(modifier = Modifier.height(10.dp))
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(Color(0x1A1F6FEB), RoundedCornerShape(10.dp))
                                    .border(1.dp, Color(0x331F6FEB), RoundedCornerShape(10.dp))
                                    .padding(10.dp)
                            ) {
                                  Text(
                                      text = stringResource(id = R.string.usb_desc),
                                      color = Color(0xFF79C0FF),
                                      fontSize = 11.sp
                                  )
                            }
                            Spacer(modifier = Modifier.height(16.dp))
                            Text(
                                text = stringResource(id = R.string.usb_instructions),
                                color = CTextSlate,
                                fontSize = 12.sp,
                                modifier = Modifier.padding(bottom = 16.dp)
                            )
                            Button(
                                onClick = { onConnect("ws://127.0.0.1:6969/ws") },
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .height(54.dp),
                                shape = RoundedCornerShape(12.dp),
                                colors = ButtonDefaults.buttonColors(
                                    containerColor = CPrimary,
                                    contentColor = Color.White
                                )
                            ) {
                                Text(stringResource(id = R.string.connect_usb), fontSize = 15.sp, fontWeight = FontWeight.Bold)
                            }
                        }

                        ConnectionMode.Bluetooth -> {
                            Text(
                                text = stringResource(id = R.string.bt_no_wifi),
                                color = CTextWhite,
                                fontSize = 15.sp,
                                fontWeight = FontWeight.Bold,
                                modifier = Modifier.align(Alignment.Start)
                            )
                            Spacer(modifier = Modifier.height(10.dp))
                            Box(
                                modifier = Modifier
                                    .fillMaxWidth()
                                    .background(Color(0x1A8B949E), RoundedCornerShape(10.dp))
                                    .border(1.dp, CBorderActive, RoundedCornerShape(10.dp))
                                    .padding(10.dp)
                            ) {
                                Text(
                                    text = stringResource(id = R.string.bt_desc),
                                    color = CTextSlate,
                                    fontSize = 11.sp
                                )
                            }
                            Spacer(modifier = Modifier.height(14.dp))

                            if (bluetoothAdapter == null) {
                                Text(
                                    text = stringResource(id = R.string.bt_not_available),
                                    color = CError,
                                    fontSize = 13.sp,
                                    textAlign = TextAlign.Center
                                )
                            } else if (!bluetoothPermissionGranted && Build.VERSION.SDK_INT >= Build.VERSION_CODES.S) {
                                Text(
                                    text = stringResource(id = R.string.bt_permission_required),
                                    color = CTextSlate,
                                    fontSize = 12.sp,
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier.padding(bottom = 12.dp)
                                )
                                Button(
                                    onClick = { permissionLauncher.launch(Manifest.permission.BLUETOOTH_CONNECT) },
                                    colors = ButtonDefaults.buttonColors(containerColor = CPrimary),
                                    shape = RoundedCornerShape(10.dp)
                                ) {
                                    Text(stringResource(id = R.string.grant_permission), color = Color.White)
                                }
                            } else {
                                Text(
                                    text = stringResource(id = R.string.paired_devices),
                                    color = CTextWhite,
                                    fontSize = 13.sp,
                                    fontWeight = FontWeight.Bold,
                                    modifier = Modifier
                                        .align(Alignment.Start)
                                        .padding(bottom = 8.dp)
                                )
                                if (pairedDevices.isEmpty()) {
                                    Text(
                                        text = stringResource(id = R.string.no_devices_found),
                                        color = Color(0xFF484F58),
                                        fontSize = 12.sp,
                                        textAlign = TextAlign.Center,
                                        modifier = Modifier.padding(vertical = 12.dp)
                                    )
                                } else {
                                    Column(
                                        modifier = Modifier
                                            .fillMaxWidth()
                                            .clip(RoundedCornerShape(12.dp))
                                            .background(CBg)
                                            .border(1.dp, CBorder, RoundedCornerShape(12.dp))
                                            .padding(4.dp)
                                    ) {
                                        pairedDevices.forEach { device ->
                                            val deviceName = try { device.name } catch (e: SecurityException) { "Device" }
                                            val deviceAddress = device.address
                                            Row(
                                                modifier = Modifier
                                                    .fillMaxWidth()
                                                    .clickable { onConnectBluetooth(deviceAddress) }
                                                    .padding(horizontal = 12.dp, vertical = 10.dp),
                                                verticalAlignment = Alignment.CenterVertically
                                            ) {
                                                Column(modifier = Modifier.weight(1f)) {
                                                    Text(
                                                        text = deviceName ?: "Unknown",
                                                        color = CTextWhite,
                                                        fontSize = 13.sp,
                                                        fontWeight = FontWeight.SemiBold
                                                    )
                                                    Text(
                                                        text = deviceAddress,
                                                        color = CTextSlate,
                                                        fontSize = 11.sp
                                                    )
                                                }
                                                Text(
                                                    text = stringResource(id = R.string.connect_action),
                                                    color = CPrimary,
                                                    fontSize = 12.sp,
                                                    fontWeight = FontWeight.Bold
                                                )
                                            }
                                        }
                                    }
                                }

                                Spacer(modifier = Modifier.height(14.dp))
                                Text(
                                    text = stringResource(id = R.string.or_enter_mac),
                                    color = Color(0xFF484F58),
                                    fontSize = 11.sp,
                                    textAlign = TextAlign.Center,
                                    modifier = Modifier.padding(vertical = 4.dp)
                                )
                                OutlinedTextField(
                                    value = btMac,
                                    onValueChange = { btMac = it },
                                    label = { Text(stringResource(id = R.string.bt_mac_label), color = CTextSlate) },
                                    placeholder = { Text(stringResource(id = R.string.bt_mac_placeholder), color = Color(0xFF484F58)) },
                                    modifier = Modifier.fillMaxWidth(),
                                    singleLine = true,
                                    colors = OutlinedTextFieldDefaults.colors(
                                        focusedBorderColor = CPrimary,
                                        unfocusedBorderColor = CBorderActive,
                                        focusedLabelColor = CPrimary,
                                        unfocusedLabelColor = CTextSlate,
                                        focusedTextColor = CTextWhite,
                                        unfocusedTextColor = CTextWhite,
                                        cursorColor = CPrimary
                                    ),
                                    shape = RoundedCornerShape(12.dp)
                                )
                                Spacer(modifier = Modifier.height(14.dp))
                                Button(
                                    onClick = { if (btMac.isNotEmpty()) onConnectBluetooth(btMac.trim()) },
                                    modifier = Modifier
                                        .fillMaxWidth()
                                        .height(50.dp),
                                    shape = RoundedCornerShape(12.dp),
                                    colors = ButtonDefaults.buttonColors(
                                        containerColor = CPrimary,
                                        contentColor = Color.White
                                    )
                                ) {
                                    Text(stringResource(id = R.string.connect_manually), fontSize = 14.sp, fontWeight = FontWeight.Bold)
                                }
                            }
                        }
                    }

                    if (errorMessage.isNotEmpty()) {
                        Spacer(modifier = Modifier.height(18.dp))
                        Text(
                            text = errorMessage,
                            color = CError,
                            fontSize = 13.sp,
                            fontWeight = FontWeight.Medium,
                            textAlign = TextAlign.Center,
                            modifier = Modifier
                                .fillMaxWidth()
                                .background(Color(0x1AE74C3C), RoundedCornerShape(10.dp))
                                .border(1.dp, Color(0x33E74C3C), RoundedCornerShape(10.dp))
                                .padding(12.dp)
                        )
                    }
                }
            }

            Spacer(modifier = Modifier.height(40.dp))
        }
    }
}
