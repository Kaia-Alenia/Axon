package com.example.axon.network

import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.WebSocket
import okhttp3.WebSocketListener
import okhttp3.Response
import org.json.JSONObject
import java.util.concurrent.TimeUnit

class WebSocketClient(private val listener: WebSocketConnectionListener) : InputClient {
    private var serverIp: String = ""

    override fun getServerIp(): String {
        return serverIp
    }

    private var client: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(5, TimeUnit.SECONDS)
        .readTimeout(0, TimeUnit.MILLISECONDS)
        .writeTimeout(0, TimeUnit.MILLISECONDS)
        .build()
    
    private var webSocket: WebSocket? = null

    interface WebSocketConnectionListener {
        fun onConnected()
        fun onDisconnected()
        fun onError(message: String)
    }

    fun connect(url: String) {
        try {
            serverIp = java.net.URI(url).host ?: ""
        } catch (e: Exception) {
            serverIp = ""
        }
        val request = Request.Builder().url(url).build()
        webSocket = client.newWebSocket(request, object : WebSocketListener() {
            override fun onOpen(webSocket: WebSocket, response: Response) {
                listener.onConnected()
            }

            override fun onMessage(webSocket: WebSocket, text: String) {
                try {
                    val json = org.json.JSONObject(text)
                    if (json.optString("type") == "ping") {
                        val timestamp = json.optLong("timestamp")
                        val responseJson = org.json.JSONObject()
                        responseJson.put("type", "pong")
                        responseJson.put("timestamp", timestamp)
                        webSocket.send(responseJson.toString())
                    }
                } catch (e: Exception) {
                }
            }

            override fun onClosed(webSocket: WebSocket, code: Int, reason: String) {
                listener.onDisconnected()
            }

            override fun onFailure(webSocket: WebSocket, t: Throwable, response: Response?) {
                listener.onError(t.message ?: "Connection Failure")
            }
        })
    }

    override fun disconnect() {
        webSocket?.close(1000, "Normal Close")
        webSocket = null
        try {
            client.connectionPool.evictAll()
        } catch (e: Exception) {
        }
    }

    fun send(message: String) {
        webSocket?.send(message)
    }

    override fun sendMove(dx: Double, dy: Double) {
        send("{\"type\":\"move\",\"dx\":$dx,\"dy\":$dy}")
    }

    override fun sendClick(button: String) {
        val json = JSONObject()
        json.put("type", "click")
        json.put("button", button)
        send(json.toString())
    }

    override fun sendMouseDown(button: String) {
        val json = JSONObject()
        json.put("type", "mousedown")
        json.put("button", button)
        send(json.toString())
    }

    override fun sendMouseUp(button: String) {
        val json = JSONObject()
        json.put("type", "mouseup")
        json.put("button", button)
        send(json.toString())
    }

    override fun sendScroll(dy: Double) {
        val json = JSONObject()
        json.put("type", "scroll")
        json.put("dy", dy)
        send(json.toString())
    }

    override fun sendType(text: String) {
        val json = JSONObject()
        json.put("type", "type")
        json.put("text", text)
        send(json.toString())
    }

    override fun sendKey(key: String) {
        val json = JSONObject()
        json.put("type", "key")
        json.put("key", key)
        send(json.toString())
    }

    override fun sendKeyCombo(modifier: String, key: String) {
        val json = JSONObject()
        json.put("type", "keycombo")
        json.put("modifier", modifier)
        json.put("key", key)
        send(json.toString())
    }
}
