(function() {
    let socket;
    let touchpad = document.getElementById("touchpad");
    let scrollpad = document.getElementById("scrollpad");
    let statusIndicator = document.getElementById("connection-status");
    let btnLeft = document.getElementById("btn-left");
    let btnRight = document.getElementById("btn-right");
    let btnKeyboard = document.getElementById("btn-keyboard");
    let hiddenInput = document.getElementById("hidden-input");

    let lastX = 0;
    let lastY = 0;
    let isMoving = false;
    let touchStartTime = 0;
    let touchStartX = 0;
    let touchStartY = 0;
    let sensitivity = 1.8;

    let lastScrollY = 0;
    let isScrolling = false;

    const translations = {
        en: {
            connected: "Connected",
            disconnected: "Disconnected",
            left: "Left",
            right: "Right",
            keyboard: "Keyboard",
            touchpad: "TOUCHPAD",
            scroll: "SCROLL"
        },
        es: {
            connected: "Conectado",
            disconnected: "Desconectado",
            left: "Izquierdo",
            right: "Derecho",
            keyboard: "Teclado",
            touchpad: "TOUCHPAD",
            scroll: "DESPLAZAR"
        }
    };

    const userLang = (navigator.language || navigator.userLanguage || "en").startsWith("es") ? "es" : "en";
    const t = translations[userLang];
    document.documentElement.lang = userLang;

    function updateStatus(isConnected) {
        if (isConnected) {
            statusIndicator.textContent = t.connected;
            statusIndicator.className = "status-indicator connected";
        } else {
            statusIndicator.textContent = t.disconnected;
            statusIndicator.className = "status-indicator disconnected";
        }
    }

    function applyTranslations() {
        btnLeft.textContent = t.left;
        btnRight.textContent = t.right;
        btnKeyboard.textContent = t.keyboard;
        
        const touchpadLabel = document.querySelector(".touchpad-label");
        if (touchpadLabel) touchpadLabel.textContent = t.touchpad;
        
        const scrollpadLabel = document.querySelector(".scrollpad-label");
        if (scrollpadLabel) scrollpadLabel.textContent = t.scroll;
        
        updateStatus(false);
    }

    function connect() {
        let proto = window.location.protocol === "https:" ? "wss:" : "ws:";
        let wsUrl = proto + "//" + window.location.host + "/ws";
        
        socket = new WebSocket(wsUrl);

        socket.onopen = function() {
            updateStatus(true);
        };

        socket.onclose = function() {
            updateStatus(false);
            setTimeout(connect, 2000);
        };

        socket.onerror = function() {
            socket.close();
        };
    }

    function sendMessage(msg) {
        if (socket && socket.readyState === WebSocket.OPEN) {
            socket.send(JSON.stringify(msg));
        }
    }

    touchpad.addEventListener("touchstart", function(e) {
        if (e.touches.length === 1) {
            let touch = e.touches[0];
            lastX = touch.clientX;
            lastY = touch.clientY;
            touchStartX = touch.clientX;
            touchStartY = touch.clientY;
            touchStartTime = Date.now();
            isMoving = true;
        }
    }, { passive: true });

    touchpad.addEventListener("touchmove", function(e) {
        if (isMoving && e.touches.length === 1) {
            let touch = e.touches[0];
            let dx = (touch.clientX - lastX) * sensitivity;
            let dy = (touch.clientY - lastY) * sensitivity;
            
            sendMessage({
                type: "move",
                dx: dx,
                dy: dy
            });

            lastX = touch.clientX;
            lastY = touch.clientY;
        }
    }, { passive: true });

    touchpad.addEventListener("touchend", function(e) {
        isMoving = false;
        let touchDuration = Date.now() - touchStartTime;
        if (e.changedTouches.length === 1 && touchDuration < 200) {
            let touch = e.changedTouches[0];
            let dist = Math.hypot(touch.clientX - touchStartX, touch.clientY - touchStartY);
            if (dist < 10) {
                sendMessage({
                    type: "click",
                    button: "left"
                });
            }
        }
    }, { passive: true });

    scrollpad.addEventListener("touchstart", function(e) {
        if (e.touches.length === 1) {
            lastScrollY = e.touches[0].clientY;
            isScrolling = true;
        }
    }, { passive: true });

    scrollpad.addEventListener("touchmove", function(e) {
        if (isScrolling && e.touches.length === 1) {
            let currentScrollY = e.touches[0].clientY;
            let diffY = currentScrollY - lastScrollY;
            
            if (Math.abs(diffY) > 8) {
                sendMessage({
                    type: "scroll",
                    dy: diffY > 0 ? 1 : -1
                });
                lastScrollY = currentScrollY;
            }
        }
    }, { passive: true });

    scrollpad.addEventListener("touchend", function() {
        isScrolling = false;
    }, { passive: true });

    btnLeft.addEventListener("touchstart", function(e) {
        e.preventDefault();
        sendMessage({ type: "mousedown", button: "left" });
    });

    btnLeft.addEventListener("touchend", function(e) {
        e.preventDefault();
        sendMessage({ type: "mouseup", button: "left" });
    });

    btnRight.addEventListener("touchstart", function(e) {
        e.preventDefault();
        sendMessage({ type: "mousedown", button: "right" });
    });

    btnRight.addEventListener("touchend", function(e) {
        e.preventDefault();
        sendMessage({ type: "mouseup", button: "right" });
    });

    btnKeyboard.addEventListener("click", function() {
        hiddenInput.focus();
    });

    hiddenInput.addEventListener("input", function(e) {
        let text = e.target.value;
        if (text.length > 0) {
            sendMessage({
                type: "type",
                text: text
            });
            e.target.value = "";
        }
    });

    hiddenInput.addEventListener("keydown", function(e) {
        if (e.key === "Backspace") {
            sendMessage({
                type: "key",
                key: "BackSpace"
            });
        } else if (e.key === "Enter") {
            sendMessage({
                type: "key",
                key: "Return"
            });
        }
    });

    applyTranslations();
    connect();
})();
