package com.example.axon.network

import org.junit.Assert.assertArrayEquals
import org.junit.Test
import java.nio.ByteBuffer

class UdpClientTest {

    @Test
    fun testCreateMovePacket() {
        val dx = 10.5
        val dy = -5.25

        val expectedBytes = ByteArray(9)
        expectedBytes[0] = 0.toByte()

        val xBits = java.lang.Float.floatToIntBits(dx.toFloat())
        expectedBytes[1] = (xBits shr 24).toByte()
        expectedBytes[2] = (xBits shr 16).toByte()
        expectedBytes[3] = (xBits shr 8).toByte()
        expectedBytes[4] = xBits.toByte()

        val yBits = java.lang.Float.floatToIntBits(dy.toFloat())
        expectedBytes[5] = (yBits shr 24).toByte()
        expectedBytes[6] = (yBits shr 16).toByte()
        expectedBytes[7] = (yBits shr 8).toByte()
        expectedBytes[8] = yBits.toByte()

        val actualBytes = UdpClient.createMovePacket(dx, dy)

        assertArrayEquals(expectedBytes, actualBytes)
    }

    @Test
    fun testCreateMovePacket_zeroValues() {
        val dx = 0.0
        val dy = 0.0

        val expectedBytes = ByteArray(9)
        expectedBytes[0] = 0.toByte()

        val xBits = java.lang.Float.floatToIntBits(0f)
        expectedBytes[1] = (xBits shr 24).toByte()
        expectedBytes[2] = (xBits shr 16).toByte()
        expectedBytes[3] = (xBits shr 8).toByte()
        expectedBytes[4] = xBits.toByte()

        val yBits = java.lang.Float.floatToIntBits(0f)
        expectedBytes[5] = (yBits shr 24).toByte()
        expectedBytes[6] = (yBits shr 16).toByte()
        expectedBytes[7] = (yBits shr 8).toByte()
        expectedBytes[8] = yBits.toByte()

        val actualBytes = UdpClient.createMovePacket(dx, dy)

        assertArrayEquals(expectedBytes, actualBytes)
    }

    @Test
    fun testCreateMovePacket_usingByteBufferToVerify() {
        val dx = 123.456
        val dy = -789.012

        val actualBytes = UdpClient.createMovePacket(dx, dy)

        val buffer = ByteBuffer.wrap(actualBytes)
        val type = buffer.get()
        val actualDx = buffer.float
        val actualDy = buffer.float

        assertArrayEquals(
            byteArrayOf(0),
            byteArrayOf(type)
        )
        // Note: double to float precision loss is expected, so we compare float values
        org.junit.Assert.assertEquals(dx.toFloat(), actualDx, 0.0001f)
        org.junit.Assert.assertEquals(dy.toFloat(), actualDy, 0.0001f)
    }
}
