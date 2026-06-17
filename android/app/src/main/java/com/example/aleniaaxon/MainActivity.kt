package com.example.aleniaaxon

import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import com.example.aleniaaxon.network.BluetoothClient
import com.example.aleniaaxon.network.InputClient
import com.example.aleniaaxon.network.WebSocketClient
import com.example.aleniaaxon.theme.AleniaAxonTheme
import com.example.aleniaaxon.ui.ConnectScreen
import com.example.aleniaaxon.ui.MainScreen
import com.example.aleniaaxon.ui.QrScannerScreen

enum class AppScreen {
    Connect,
    Scan,
    Control
}

class MainActivity : ComponentActivity(), WebSocketClient.WebSocketConnectionListener, BluetoothClient.ConnectionListener {
    private var webSocketClient: WebSocketClient? = null
    private var bluetoothClient: BluetoothClient? = null
    private var activeClient: InputClient? = null
    private var currentScreen by mutableStateOf(AppScreen.Connect)
    private var connectionError by mutableStateOf("")

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        webSocketClient = WebSocketClient(this)
        bluetoothClient = BluetoothClient(this)
        
        enableEdgeToEdge()
        setContent {
            AleniaAxonTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    when (currentScreen) {
                        AppScreen.Connect -> {
                            ConnectScreen(
                                onConnect = { url ->
                                    connectionError = ""
                                    activeClient = webSocketClient
                                    webSocketClient?.connect(url)
                                },
                                onConnectBluetooth = { mac ->
                                    connectionError = ""
                                    activeClient = bluetoothClient
                                    bluetoothClient?.connect(mac)
                                },
                                onScanQr = {
                                    currentScreen = AppScreen.Scan
                                },
                                errorMessage = connectionError
                            )
                        }
                        AppScreen.Scan -> {
                            QrScannerScreen(
                                onQrDetected = { url ->
                                    currentScreen = AppScreen.Connect
                                    connectionError = ""
                                    val wsUrl = try {
                                        val uri = java.net.URI(url)
                                        val query = if (uri.rawQuery != null) "?${uri.rawQuery}" else ""
                                        "ws://${uri.host}:${uri.port}/ws$query"
                                    } catch (e: Exception) {
                                        url.replace("http://", "ws://").replace("/?", "/ws?").replace("http://", "ws://")
                                    }
                                    activeClient = webSocketClient
                                    webSocketClient?.connect(wsUrl)
                                },
                                onBack = {
                                    currentScreen = AppScreen.Connect
                                }
                            )
                        }
                        AppScreen.Control -> {
                            activeClient?.let { client ->
                                MainScreen(
                                    client = client,
                                    onDisconnect = {
                                        client.disconnect()
                                        currentScreen = AppScreen.Connect
                                    }
                                )
                            }
                        }
                    }
                }
            }
        }
    }

    override fun onConnected() {
        runOnUiThread {
            currentScreen = AppScreen.Control
            connectionError = ""
        }
    }

    override fun onDisconnected() {
        runOnUiThread {
            currentScreen = AppScreen.Connect
        }
    }

    override fun onError(message: String) {
        runOnUiThread {
            currentScreen = AppScreen.Connect
            connectionError = message
        }
    }

    override fun onDestroy() {
        super.onDestroy()
        webSocketClient?.disconnect()
        bluetoothClient?.disconnect()
    }
}
