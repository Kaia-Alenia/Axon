(function() {
    let socket;
    let touchpad = document.getElementById("touchpad");
    let scrollpad = document.getElementById("scrollpad");
    let statusIndicator = document.getElementById("connection-status");
    let btnLeft = document.getElementById("btn-left");
    let btnRight = document.getElementById("btn-right");
    let btnKeyboard = document.getElementById("btn-keyboard");
    let hiddenInput = document.getElementById("hidden-input");
    let btnFnToggle = document.getElementById("btn-fn-toggle");
    let fnContainer = document.getElementById("fn-container");
    let lastX = 0;
    let lastY = 0;
    let isMoving = false;
    let touchStartTime = 0;
    let touchStartX = 0;
    let touchStartY = 0;
    let sensitivity = 1.8;
    let lastScrollY = 0;
    let isScrolling = false;
    let twoFingerStartY = 0;
    let lastTwoFingerY = 0;
    let lastTwoFingerDist = 0;
    let twoFingerScrollAcc = 0;
    let isTwoFingerGesture = false;
    const translations = {
        en: {
            connected: "Connected",
            disconnected: "Disconnected",
            left: "Left",
            right: "Right",
            keyboard: "Keyboard",
            touchpad: "TOUCHPAD",
            scroll: "SCROLL",
            prevVideo: "Prev",
            nextVideo: "Next"
        },
        es: {
            connected: "Conectado",
            disconnected: "Desconectado",
            left: "Izquierdo",
            right: "Derecho",
            keyboard: "Teclado",
            touchpad: "TOUCHPAD",
            scroll: "DESPLAZAR",
            prevVideo: "Anterior",
            nextVideo: "Siguiente"
        },
        zh: {
            connected: "已连接",
            disconnected: "已断开",
            left: "左键",
            right: "右键",
            keyboard: "键盘",
            touchpad: "触控板",
            scroll: "滚动",
            prevVideo: "上一个",
            nextVideo: "下一个"
        }
    };
    const rawLang = navigator.language || navigator.userLanguage || "en";
    let userLang = "en";
    if (rawLang.startsWith("zh")) {
        userLang = "zh";
    } else if (rawLang.startsWith("es")) {
        userLang = "es";
    }
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
        const prevText = document.querySelector("#btn-prev-video .btn-text");
        if (prevText) prevText.textContent = t.prevVideo;
        const nextText = document.querySelector("#btn-next-video .btn-text");
        if (nextText) nextText.textContent = t.nextVideo;
        updateStatus(false);
    }
    function connect() {
        let proto = window.location.protocol === "https:" ? "wss:" : "ws:";
        let urlParams = new URLSearchParams(window.location.search);
        let token = urlParams.get("token") || "";
        let wsUrl = proto + "//" + window.location.host + "/ws" + (token ? "?token=" + token : "");
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
            isTwoFingerGesture = false;
        } else if (e.touches.length === 2) {
            e.preventDefault();
            let t1 = e.touches[0];
            let t2 = e.touches[1];
            isMoving = false;
            isTwoFingerGesture = true;
            lastTwoFingerY = (t1.clientY + t2.clientY) / 2;
            twoFingerStartY = lastTwoFingerY;
            lastTwoFingerDist = Math.hypot(t1.clientX - t2.clientX, t1.clientY - t2.clientY);
            twoFingerScrollAcc = 0;
        }
    }, { passive: false });
    touchpad.addEventListener("touchmove", function(e) {
        if (isTwoFingerGesture && e.touches.length === 2) {
            e.preventDefault();
            let t1 = e.touches[0];
            let t2 = e.touches[1];
            let currentY = (t1.clientY + t2.clientY) / 2;
            let currentDist = Math.hypot(t1.clientX - t2.clientX, t1.clientY - t2.clientY);
            let dy = currentY - lastTwoFingerY;
            let dDist = currentDist - lastTwoFingerDist;
            if (Math.abs(dy) > Math.abs(dDist) * 1.1 || Math.abs(dDist) < 10) {
                twoFingerScrollAcc += dy;
                if (Math.abs(twoFingerScrollAcc) > 12) {
                    sendMessage({
                        type: "scroll",
                        dy: twoFingerScrollAcc > 0 ? 1 : -1
                    });
                    twoFingerScrollAcc = 0;
                }
                lastTwoFingerY = currentY;
                lastTwoFingerDist = currentDist;
            } else if (Math.abs(dDist) > 20) {
                sendMessage({
                    type: "keycombo",
                    modifier: "ctrl",
                    key: dDist > 0 ? "+" : "-"
                });
                lastTwoFingerDist = currentDist;
                lastTwoFingerY = currentY;
                twoFingerScrollAcc = 0;
            }
            return;
        }
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
    }, { passive: false });
    touchpad.addEventListener("touchend", function(e) {
        if (isTwoFingerGesture) {
            if (e.touches.length < 2) {
                isTwoFingerGesture = false;
                twoFingerScrollAcc = 0;
            }
            return;
        }
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
    document.querySelectorAll(".quick-btn, .fn-btn").forEach(function(btn) {
        btn.addEventListener("click", function(e) {
            e.preventDefault();
            let key = btn.getAttribute("data-key");
            if (key) {
                sendMessage({
                    type: "key",
                    key: key
                });
            }
        });
    });
    if (btnFnToggle && fnContainer) {
        btnFnToggle.addEventListener("click", function() {
            fnContainer.classList.toggle("hidden");
            btnFnToggle.classList.toggle("active");
        });
    }
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
