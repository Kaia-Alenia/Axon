# Axon Roadmap

Este documento describe la visión a futuro y los planes de desarrollo para Axon. El objetivo es mantener una guía clara de las características planificadas, desde mejoras inmediatas hasta grandes saltos de plataforma.

## 🚀 Corto Plazo (Mejoras Inmediatas y Estabilidad)

- **Seguridad y Privacidad:**
  - [ ] Implementar un sistema de PIN/Contraseña para emparejamiento seguro (evitar conexiones no deseadas en redes públicas).
  - [ ] Cifrado de extremo a extremo (E2EE) para la transmisión de las pulsaciones de teclado.
- **Mejoras de Interfaz (Cliente):**
  - [ ] Soporte completo para temas dinámicos (Material You) y modo oscuro/claro manual.
  - [ ] Perfiles personalizados de sensibilidad y aceleración para el touchpad.
- **Nuevos Modos de Control:**
  - [ ] **Modo Presentación:** Interfaz simplificada con botones gigantes de siguiente/anterior para diapositivas (PowerPoint/Keynote) y un puntero láser virtual.
  - [ ] **Controles Multimedia Dedicados:** Pantalla específica para controlar reproducción (Play/Pause, saltar pista, controles de volumen avanzados).

## ⚡ Mediano Plazo (Evolución y Expansión)

- **Soporte Multiplataforma para el Cliente:**
  - [ ] **Cliente iOS:** Desarrollar la versión para iPhone/iPad (considerando Kotlin Multiplatform o Swift nativo).
- **Interfaz Gráfica para el Servidor:**
  - [ ] **Aplicación de Bandeja de Sistema (System Tray):** Crear una interfaz gráfica minimalista para el servidor en PC (Windows/Mac/Linux) para encender/apagar el servidor, ver dispositivos conectados y configurar puertos sin usar la terminal.
- **Inputs Avanzados:**
  - [ ] **Modo Gamepad (Giroscopio):** Usar los sensores de movimiento del teléfono (giroscopio/acelerómetro) para controlar juegos de conducción o simuladores de vuelo en la PC.
  - [ ] **Dictado por Voz:** Usar el reconocimiento de voz del teléfono para escribir texto directamente en la PC de forma ultrarrápida.

## 🔭 Largo Plazo (Visión a Futuro)

- **Expansión a Smart TVs (Android TV / Google TV):**
  - [ ] **Axon TV Server:** Adaptar el servidor para que se ejecute de forma nativa en televisores con Google TV.
  - [ ] Usar la API de *AccessibilityService* para inyectar movimientos de ratón y clics en la interfaz de la TV.
  - [ ] Permitir buscar películas o introducir contraseñas largas en la TV cómodamente desde el teclado del móvil.
- **Screen Mirroring Ligero:**
  - [ ] Transmitir la pantalla de la PC al teléfono a muy bajos FPS (para no afectar la latencia) de modo que se pueda controlar la PC incluso si el monitor está apagado o en otra habitación.
- **Integración con Automatizaciones:**
  - [ ] Atajos o *Intents* para permitir que apps como Tasker o atajos de Siri puedan ejecutar macros en la PC (ej. "Apagar PC" desde un comando de voz al celular).
