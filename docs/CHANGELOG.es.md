# Registro de Cambios

Todos los cambios notables de este proyecto se documentan en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es/1.1.0/) y este proyecto sigue el [Versionado Semántico](https://semver.org/lang/es/).


## [1.6.8] - 2026-11-04

### ✔️ Corregido
- **Bloqueo por Modo de Edición Rápida (QuickEdit) en Consola de Windows 10 y Logger No Bloqueante:** Desactivado `ENABLE_QUICK_EDIT_MODE` en `windows.Stdin` e implementado un logger de consola asíncrono para que hacer clic en la terminal de Windows nunca congele el bucle WebSocket de clics y teclado.
- **Layout Binario Exacto de 40 Bytes para `SendInput` en Windows 10:** Unificada la estructura `INPUT` de ratón y teclado en un arreglo binario exacto de 40 bytes (`rawInput [40]byte`) y envío atómico de eventos de teclado y clics.

## [1.6.7] - 2026-11-04

### ✔️ Corregido
- **Simulación de Entrada en Windows 10 (`SendInput`):** Corregida la alineación de memoria de la estructura `INPUT` de 64 bits en `simulator_windows.go` (40 bytes en lugar de 48 bytes), restaurando movimiento de ratón, clics, desplazamiento y zoom en Windows.
- **Bluetooth RFCOMM Nativo en Windows:** Implementado socket RFCOMM real sobre Winsock2 (`AF_BTH`) y registro de servicio SDP (`00001101-0000-1000-8000-00805F9B34FB`) en `bluetooth_windows.go`.
- **Detección de IP y Wi-Fi en Windows:** Añadida detección de ruta LAN activa, filtrado de adaptadores virtuales (`vEthernet`, `WSL`, `VirtualBox`, `VMware`, `Docker`) y desactivación de `SO_REUSEADDR` en Windows.

## [1.6.6] - 2026-08-23

### 🚀 Añadido / Mejorado
- Cumplimiento de estándares de código de Alenia Studios (Código Limpio).
- Actualización de versión interna en Go y Android.

## [1.6.3] - 2026-06-29

### ✔️ Corregido
- **Advertencia ADB:** Restaurado un mensaje de advertencia limpio y de una sola línea. La salida de error del sistema se suprime; solo se muestra un aviso conciso de `exit status 1` cuando ADB está instalado pero no hay dispositivo conectado.
- **Simulación de Entrada en macOS:** Separado el simulador de entrada específico de Linux (uinput/xdotool) de macOS. macOS ahora usa su propio simulador nativo respaldado por `cliclick` (instalable vía Homebrew), evitando fallos en Apple Silicon y Macs Intel.

### ✅ Añadido
- **Sección de Solución de Problemas (README):** Añadida una explicación clara de la advertencia de inicio de ADB con instrucciones de instalación específicas por plataforma.

## [1.6.2] - 2026-06-29

### ✔️ Corregido
- **Advertencias ADB:** Silenciadas las advertencias verbosas de conectividad ADB y evitado invocar ADB si no está instalado en el PATH del sistema.

## [1.6.1] - 2026-06-29

### ✔️ Corregido
- **Renderizado de Terminal en Windows:** Habilitado el Procesamiento de Terminal Virtual para Windows CMD/PowerShell para renderizar correctamente los colores ANSI y el arte ASCII.

## [1.6.0] - 2026-06-29

### ✔️ Corregido
- **Compilación en Windows:** Resuelto un error crítico de compilación multiplataforma relacionado con `SO_REUSEADDR` implementando lógica de control de socket específica por plataforma.

### ✅ Añadido

**Conectividad y Entrada**
- **Bluetooth de Baja Latencia:** Soporte Bluetooth ultrarrápido y de baja latencia optimizado específicamente para Windows y macOS.
- **Gestos Avanzados:** Nuevos gestos intuitivos para zoom.
- **Integración de Hardware:** El volumen ahora se puede controlar directamente con los botones físicos de volumen.
- **Auto-Conexión:** Emparejamiento fluido vía código QR.
- **Conectividad Versátil:** Soporte completo para modos Wi-Fi y Bluetooth.

**Arquitectura y Clientes**
- **Servidor Go:** Backend robusto utilizando protocolos WebSocket + UDP para transmisión de entrada de ultra baja latencia.
- **Cliente Web Embebido:** Soporte completo de touchpad, scroll, teclado y botones del ratón desde el navegador.
- **Cliente Android Nativo:** Construido desde cero con Jetpack Compose para una experiencia nativa y fluida.
- **Simulación de Entrada:** Simulación de entrada multiplataforma confiable en Linux, Windows y macOS.

**Proyecto e Infraestructura**
- **Pipeline CI/CD:** Flujos de compilación y publicación automatizados con GitHub Actions.
- **Estándares de Comunidad:** Añadidas guía de contribución, código de conducta y política de seguridad.
