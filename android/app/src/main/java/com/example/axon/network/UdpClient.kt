package com.example.axon.network

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
    private val scope = CoroutineScope(Dispatchers.IO)
    private val channel = Channel<ByteArray>(Channel.CONFLATED)

    init {
        scope.launch {
            try {
                val address = InetAddress.getByName(serverIp)
                val socket = DatagramSocket()
                for (bytes in channel) {
                    try {
                        val packet = DatagramPacket(bytes, bytes.size, address, port)
                        socket.send(packet)
                    } catch (e: Exception) {
                    }
                }
                socket.close()
            } catch (e: Exception) {
            }
        }
    }

    fun sendMove(dx: Double, dy: Double) {
        val bytes = ByteArray(9)
        bytes[0] = 0.toByte()
        
        val xBits = java.lang.Float.floatToIntBits(dx.toFloat())
        bytes[1] = (xBits shr 24).toByte()
        bytes[2] = (xBits shr 16).toByte()
        bytes[3] = (xBits shr 8).toByte()
        bytes[4] = xBits.toByte()

        val yBits = java.lang.Float.floatToIntBits(dy.toFloat())
        bytes[5] = (yBits shr 24).toByte()
        bytes[6] = (yBits shr 16).toByte()
        bytes[7] = (yBits shr 8).toByte()
        bytes[8] = yBits.toByte()

        channel.trySend(bytes)
    }

    fun sendScroll(dy: Double) {
        val bytes = ByteArray(9)
        bytes[0] = 1.toByte()

        val yBits = java.lang.Float.floatToIntBits(dy.toFloat())
        bytes[1] = (yBits shr 24).toByte()
        bytes[2] = (yBits shr 16).toByte()
        bytes[3] = (yBits shr 8).toByte()
        bytes[4] = yBits.toByte()

        val zeroBits = java.lang.Float.floatToIntBits(0f)
        bytes[5] = (zeroBits shr 24).toByte()
        bytes[6] = (zeroBits shr 16).toByte()
        bytes[7] = (zeroBits shr 8).toByte()
        bytes[8] = zeroBits.toByte()

        channel.trySend(bytes)
    }

    fun close() {
        try {
            channel.close()
            scope.cancel()
        } catch (e: Exception) {
        }
    }
}
