package com.example.axon.network

import android.annotation.SuppressLint
import android.bluetooth.BluetoothAdapter
import android.bluetooth.BluetoothDevice
import android.bluetooth.BluetoothSocket
import org.json.JSONObject
import java.io.OutputStream
import java.util.UUID
import java.util.concurrent.Executors

class BluetoothClient(private val listener: ConnectionListener) : InputClient {
    private var socket: BluetoothSocket? = null
    private var outputStream: OutputStream? = null
    private var connectedDeviceAddress: String = ""
    private val sendExecutor = Executors.newSingleThreadExecutor()

    interface ConnectionListener {
        fun onConnected()
        fun onDisconnected()
        fun onError(message: String)
    }

    override fun getServerIp(): String = connectedDeviceAddress

    @SuppressLint("MissingPermission")
    fun connect(deviceAddress: String) {
        connectedDeviceAddress = deviceAddress
        Thread {
            try {
                val adapter = BluetoothAdapter.getDefaultAdapter()
                val device: BluetoothDevice = adapter.getRemoteDevice(deviceAddress)
                try {
                    val uuid = UUID.fromString("00001101-0000-1000-8000-00805F9B34FB")
                    socket = device.createRfcommSocketToServiceRecord(uuid)
                    socket?.connect()
                } catch (e: Exception) {
                    try {
                        val m = device.javaClass.getMethod("createRfcommSocket", Int::class.javaPrimitiveType)
                        socket = m.invoke(device, 1) as BluetoothSocket
                        socket?.connect()
                    } catch (e2: Exception) {
                        val m = device.javaClass.getMethod("createInsecureRfcommSocket", Int::class.javaPrimitiveType)
                        socket = m.invoke(device, 1) as BluetoothSocket
                        socket?.connect()
                    }
                }
                outputStream = socket?.outputStream
                listener.onConnected()
            } catch (e: Exception) {
                listener.onError("Bluetooth Error: ${e.message}")
                disconnect()
            }
        }.start()
    }

    override fun disconnect() {
        try {
            socket?.close()
        } catch (e: Exception) {}
        socket = null
        outputStream = null
        listener.onDisconnected()
    }

    private fun send(message: String) {
        sendExecutor.execute {
            try {
                outputStream?.write((message + "\n").toByteArray())
                outputStream?.flush()
            } catch (e: Exception) {
                disconnect()
            }
        }
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
