package com.example.aleniaaxon.network

import kotlinx.coroutines.CoroutineScope
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.cancel
import kotlinx.coroutines.channels.Channel
import kotlinx.coroutines.launch
import java.net.DatagramPacket
import java.net.DatagramSocket
import java.net.InetAddress

class UdpClient(
    private val serverIp: String,
    private val port: Int = 6970
) {
    private val socket: DatagramSocket? = try {
        DatagramSocket()
    } catch (e: Exception) {
        null
    }

    private val serverAddress: InetAddress? = try {
        InetAddress.getByName(serverIp)
    } catch (e: Exception) {
        null
    }

    private val scope = CoroutineScope(Dispatchers.IO)
    private val channel = Channel<String>(Channel.CONFLATED)

    init {
        scope.launch {
            val address = serverAddress ?: return@launch
            val sock = socket ?: return@launch
            try {
                for (message in channel) {
                    try {
                        val bytes = message.toByteArray()
                        val packet = DatagramPacket(bytes, bytes.size, address, port)
                        sock.send(packet)
                    } catch (e: Exception) {
                    }
                }
            } catch (e: Exception) {
            }
        }
    }

    fun sendMove(dx: Double, dy: Double) {
        sendMessage("m:$dx:$dy")
    }

    fun sendScroll(dy: Double) {
        sendMessage("s:$dy")
    }

    private fun sendMessage(message: String) {
        channel.trySend(message)
    }

    fun close() {
        try {
            channel.close()
            socket?.close()
            scope.cancel()
        } catch (e: Exception) {
        }
    }
}
