package com.example.axon

import android.content.Context
import android.os.Bundle
import androidx.activity.ComponentActivity
import androidx.activity.compose.setContent
import androidx.activity.enableEdgeToEdge
import androidx.compose.animation.Crossfade
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.material3.MaterialTheme
import androidx.compose.material3.Surface
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.setValue
import androidx.compose.ui.Modifier
import androidx.compose.runtime.CompositionLocalProvider
import androidx.compose.runtime.remember
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.platform.LocalConfiguration
import android.content.res.Configuration
import java.util.Locale
import com.example.axon.network.BluetoothClient
import com.example.axon.network.InputClient
import com.example.axon.network.WebSocketClient
import com.example.axon.theme.AxonTheme
import com.example.axon.ui.ConnectScreen
import com.example.axon.ui.MainScreen
import com.example.axon.ui.QrScannerScreen
import com.example.axon.ui.StartScreen

enum class AppScreen {
    Start,
    Connect,
    Scan,
    Control
}

class MainActivity : ComponentActivity(), WebSocketClient.WebSocketConnectionListener, BluetoothClient.ConnectionListener {
    private var webSocketClient: WebSocketClient? = null
    private var bluetoothClient: BluetoothClient? = null
    private var activeClient: InputClient? = null
    private var currentScreen by mutableStateOf(AppScreen.Start)
    private var connectionError by mutableStateOf("")

    override fun attachBaseContext(newBase: Context) {
        val prefs = newBase.getSharedPreferences("axon_prefs", MODE_PRIVATE)
        val lang = prefs.getString("language", Locale.getDefault().language) ?: "es"
        val locale = Locale(lang)
        Locale.setDefault(locale)
        val config = Configuration(newBase.resources.configuration)
        config.setLocale(locale)
        val context = newBase.createConfigurationContext(config)
        super.attachBaseContext(context)
    }

    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        
        webSocketClient = WebSocketClient(this)
        bluetoothClient = BluetoothClient(this)
        
        enableEdgeToEdge()
        setContent {
            val prefs = remember { getSharedPreferences("axon_prefs", MODE_PRIVATE) }
            val currentLanguage = remember { prefs.getString("language", Locale.getDefault().language) ?: "es" }

            AxonTheme {
                Surface(
                    modifier = Modifier.fillMaxSize(),
                    color = MaterialTheme.colorScheme.background
                ) {
                    Crossfade(targetState = currentScreen, label = "screenTransition") { screen ->
                        when (screen) {
                            AppScreen.Start -> {
                                StartScreen(
                                    onStart = {
                                        currentScreen = AppScreen.Connect
                                    },
                                    currentLanguage = currentLanguage,
                                    onToggleLanguage = {
                                        val newLang = if (currentLanguage == "es") "en" else "es"
                                        prefs.edit().putString("language", newLang).apply()
                                        recreate()
                                    }
                                )
                            }
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
                                    onBack = {
                                        currentScreen = AppScreen.Start
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
                                            currentScreen = AppScreen.Start
                                        }
                                    )
                                }
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
