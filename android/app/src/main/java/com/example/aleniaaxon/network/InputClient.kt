package com.example.aleniaaxon.network

interface InputClient {
    fun sendMove(dx: Double, dy: Double)
    fun sendScroll(dy: Double)
    fun sendClick(button: String)
    fun sendMouseDown(button: String)
    fun sendMouseUp(button: String)
    fun sendType(text: String)
    fun sendKey(key: String)
    fun sendKeyCombo(modifier: String, key: String)
    fun disconnect()
    fun getServerIp(): String
}
