package com.example.axon.network

import org.junit.Assert.assertEquals
import org.junit.Test

class WebSocketClientTest {
    @Test
    fun connect_withInvalidUrl_resetsServerIp() {
        val listener = object : WebSocketClient.WebSocketConnectionListener {
            override fun onConnected() {}
            override fun onDisconnected() {}
            override fun onError(message: String) {}
        }
        val client = WebSocketClient(listener)

        try {
            client.connect("http://192.168.1.100:8080")
        } catch (e: Exception) {}
        assertEquals("192.168.1.100", client.getServerIp())

        try {
            client.connect("invalid://^")
        } catch (e: Exception) {}

        assertEquals("", client.getServerIp())
    }
}
