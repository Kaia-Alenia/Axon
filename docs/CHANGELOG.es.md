# Registro de Cambios

Todos los cambios notables de este proyecto se documentan en este archivo.

El formato está basado en [Keep a Changelog](https://keepachangelog.com/es/1.1.0/) y este proyecto sigue el [Versionado Semántico](https://semver.org/lang/es/).


## [1.7.7] - 2026-11-16

### Corregido
- **Diseno del encabezado del touchpad:** los dos botones con texto largo anadidos en 1.7.5 saturaban la fila del encabezado en telefonos estrechos y aplastaban el titulo Touchpad + la IP a ancho cero, asi que "Touchpad" se veia una letra por linea bajando por el lado izquierdo. El encabezado vuelve a su diseno original (titulo + IP + desconectar); los botones de la tira de desplazamiento y pantalla completa se movieron a la fila de acciones rapidas (que se desplaza horizontalmente) con etiquetas cortas. Los botones Anterior/Siguiente vuelven a sus etiquetas originales (se revierte el confuso cambio a "Video arriba/abajo").

## [1.7.6] - 2026-11-16

### Añadido
- **Nuevo logo de Axon y banner del README:** el icono de la app Android ahora es el nuevo logo del ratón con neurona (degradado cian a morado sobre fondo azul oscuro) y el README abre con el nuevo banner de Axon ("REMOTE MOUSE"). Los archivos fuente del logo están en `docs/` (`axon-logo.svg` para el icono de la app, `axon-banner.svg` para el README). El icono adaptable es una conversión vectorial del logo (`ic_launcher_foreground.xml`), y los iconos mipmap clásicos se regeneraron desde él en todas las densidades.

## [Sin publicar]

### ✔️ Corregido
- **No se podía volver al chino al cambiar de idioma (#11, bug):** el selector de la pantalla inicial solo alternaba ES <-> EN, así que al salir del chino ya no se podía volver. El selector ahora muestra los tres idiomas (EN / ES / 中文) y cada uno se puede elegir directamente; el chino usa `Locale.SIMPLIFIED_CHINESE` (zh-Hans-CN) para que los recursos `values-zh` coincidan, y las etiquetas guardadas como `zh-CN` se normalizan al cargar.
- **La tecla de borrar solo borraba lo escrito en el teléfono (#11):** el IME se tragaba la tecla de borrar cuando el buffer local estaba vacío y nada llegaba a la PC. Ahora se intercepta (`onKeyEvent`) en el touchpad y en el teclado y siempre envía un evento Backspace real a la PC, y ambas pantallas tienen un botón ⌫ dedicado.
- **Código QR muy pequeño y de colores (#12):** el servidor imprimía el QR con un degradado arcoíris ANSI y en tamaño pequeño. Ahora imprime un QR grande en blanco y negro (módulos a tamaño completo, zona de silencio de 4 módulos, sin códigos de color), y la pantalla de escaneo tiene un deslizador de zoom (zoom lineal de CameraX) con un marco que crece con el zoom.
- **FastGithub rompe el mouse (#12):** el simulador de Windows ahora revisa el valor de retorno de `SendInput` y registra una advertencia cuando Windows rechaza la entrada sintética. La causa habitual es UIPI: la app en primer plano (por ejemplo FastGithub) corre como Administrador mientras Axon no, y Windows descarta la entrada en silencio. El log del servidor ahora indica ejecutar `axon.exe` como Administrador. Esto no se pudo reproducir en esta VM Linux, así que el comportamiento en Windows está diagnosticado, no verificado en Windows.
- **Teclas de función y multimedia descartadas en silencio en Linux/macOS (#9, #10):** en Linux, las teclas que el mapa uinput no cubría (F1–F4, F6–F12, silencio, reproducir/pausa) se perdían cuando el teclado uinput estaba disponible; ahora usan xdotool como respaldo. En macOS, F1–F12 se mapean a los nombres en minúscula de cliclick y las teclas de volumen se manejan con AppleScript (volumen relativo, silencio, F8 para reproducir/pausa). Windows ya manejaba todas estas teclas.

### 🚀 Añadido / Mejorado
- **Desplazamiento con dos dedos vs zoom con pellizco (#9):** el touchpad ahora clasifica el gesto de dos dedos una sola vez — desplazar hace scroll, solo un pellizco real hace zoom (Ctrl + / Ctrl −) — y mantiene esa clasificación hasta levantar los dedos, así un scroll ya no se convierte en zoom a mitad de camino. Arrastrar con dos dedos ya no mueve el cursor del mouse.
- **Toque con dos dedos = clic derecho (#11):** un toque rápido con dos dedos en el touchpad envía un clic derecho; la pista del gesto se muestra en el touchpad.
- **Tira de desplazamiento ocultable y touchpad a pantalla completa (#11):** el encabezado del touchpad tiene interruptores para ocultar la tira de desplazamiento y para pasar a modo touchpad a pantalla completa (toda la pantalla es el touchpad, sin otros botones, con botón para salir).
- **Botones de volumen y de video de Douyin (#10):** las filas multimedia del touchpad y del teclado ya tenían volumen +/−/silencio y botones de video Arriba/Abajo, y las teclas físicas de volumen del teléfono ya enviaban eventos de volumen; los botones de video ahora dicen "Video ↑ / Video ↓" (EN/ES) / "上一个视频 / 下一个视频" (ZH) para que su propósito en Douyin quede claro, y el servidor ahora sí los entrega en Linux y macOS (ver Corregido).
- **La autoactualización del servidor por fin reemplaza el binario:** `axon --update` descargaba el binario nuevo como `axon.exe.new` y dejaba al usuario borrar y renombrar a mano — en Windows el renombre siempre fallaba porque el exe en ejecución está bloqueado. Ahora consulta la última versión en la API de GitHub (y no descarga nada si ya está actualizado), verifica la descarga (bytes contra Content-Length y el tamaño del asset de la release, más los bytes mágicos), comprueba antes el espacio libre y que el directorio sea escribible, y reemplaza el binario con respaldo (`axon.exe.old`): en Windows un proceso ayudante espera a que el padre salga, intercambia los archivos, verifica que el binario nuevo arranca de verdad (`--version` imprime la versión nueva) y solo entonces borra el respaldo — si el intercambio o la verificación fallan, restaura el binario viejo desde el respaldo. Cada paso queda en `update.log` junto al exe. En Linux/macOS el intercambio es directo y el proceso se reinicia con el binario nuevo.
- **Actualizador dentro de la app Android:** la app no tenía actualizador — nunca buscaba una versión nueva. Ahora consulta la API de releases de GitHub una vez por inicio (y a pedido con el botón "Check for updates"), muestra un diálogo dentro de la app con la versión nueva, descarga el APK con progreso a una ubicación estable (`filesDir/updates/`), verifica su tamaño y bytes ZIP antes de instalar, y lo instala con el instalador del sistema vía FileProvider — sin abrir GitHub en el navegador. Maneja el permiso "Instalar apps desconocidas" (lleva a Ajustes y reintenta al volver), instalaciones canceladas por el usuario (se cierra en silencio, el archivo se guarda para reintentar después) y descargas incompletas o corruptas (se borran, con reintento). El APK descargado solo se borra cuando se confirma que la versión nueva está instalada.

## [1.6.9] - 2026-11-05

### ✔️ Corregido
- **Ejecución en Segundo Plano al Minimizar la Ventana en Windows 10:** Configurado `SetProcessInformation` para desactivar Power Throttling / EcoQoS, elevado a `ABOVE_NORMAL_PRIORITY_CLASS` y `SetThreadExecutionState` para evitar suspensión o degradación de hilos de red (WebSocket / UDP / Bluetooth) por el kernel de Windows.
- **Protección contra Bloqueo y Crashes de Buffer en la Consola:** Detección de ventana minimizada con `IsIconic` para suspender escrituras a `os.Stdout` mientras la ventana está minimizada (evitando errores de buffer reflow y cierres de `conhost.exe`). Desactivados `ENABLE_WINDOW_INPUT` y `ENABLE_MOUSE_INPUT` en la consola.
- **Tolerancia y Resiliencia de Conexión:** Aumentado el plazo de escritura en el bucle de ping de WebSocket de 3s a 10s para prevenir desconexiones durante transiciones de ventana. Añadido parámetro `--minimized` para iniciar directamente minimizado.

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
