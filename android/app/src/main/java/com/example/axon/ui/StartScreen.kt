package com.example.axon.ui
import androidx.compose.animation.core.Animatable
import androidx.compose.animation.core.EaseOutCubic
import androidx.compose.animation.core.LinearEasing
import androidx.compose.animation.core.RepeatMode
import androidx.compose.animation.core.animateFloat
import androidx.compose.animation.core.infiniteRepeatable
import androidx.compose.animation.core.rememberInfiniteTransition
import androidx.compose.animation.core.tween
import androidx.compose.foundation.Canvas
import androidx.compose.foundation.background
import androidx.compose.foundation.border
import androidx.compose.foundation.clickable
import androidx.compose.foundation.layout.Arrangement
import androidx.compose.foundation.layout.Box
import androidx.compose.foundation.layout.Column
import androidx.compose.foundation.layout.PaddingValues
import androidx.compose.foundation.layout.Row
import androidx.compose.foundation.layout.Spacer
import androidx.compose.foundation.layout.fillMaxHeight
import androidx.compose.foundation.layout.fillMaxSize
import androidx.compose.foundation.layout.fillMaxWidth
import androidx.compose.foundation.layout.height
import androidx.compose.foundation.layout.padding
import androidx.compose.foundation.layout.size
import androidx.compose.foundation.layout.width
import androidx.compose.foundation.rememberScrollState
import androidx.compose.foundation.shape.RoundedCornerShape
import androidx.compose.foundation.verticalScroll
import androidx.compose.material3.Button
import androidx.compose.material3.ButtonDefaults
import androidx.compose.material3.Card
import androidx.compose.material3.CardDefaults
import androidx.compose.material3.LinearProgressIndicator
import androidx.compose.material3.Text
import androidx.compose.material3.TextButton
import androidx.activity.compose.rememberLauncherForActivityResult
import androidx.activity.result.contract.ActivityResultContracts
import androidx.compose.runtime.Composable
import androidx.compose.runtime.DisposableEffect
import androidx.compose.runtime.LaunchedEffect
import androidx.compose.runtime.getValue
import androidx.compose.runtime.mutableStateOf
import androidx.compose.runtime.remember
import androidx.compose.runtime.rememberCoroutineScope
import androidx.compose.runtime.setValue
import androidx.lifecycle.Lifecycle
import androidx.lifecycle.LifecycleEventObserver
import androidx.lifecycle.compose.LocalLifecycleOwner
import androidx.compose.ui.Alignment
import androidx.compose.ui.Modifier
import androidx.compose.ui.draw.alpha
import androidx.compose.ui.draw.clip
import androidx.compose.ui.geometry.Offset
import androidx.compose.ui.graphics.Brush
import androidx.compose.ui.graphics.Color
import androidx.compose.ui.graphics.Path
import androidx.compose.ui.graphics.StrokeCap
import androidx.compose.ui.graphics.drawscope.Stroke
import androidx.compose.ui.graphics.graphicsLayer
import androidx.compose.ui.platform.LocalContext
import androidx.compose.ui.res.stringResource
import androidx.compose.ui.text.font.FontWeight
import androidx.compose.ui.text.style.TextAlign
import androidx.compose.ui.unit.dp
import androidx.compose.ui.unit.sp
import androidx.compose.ui.window.Dialog
import androidx.compose.ui.window.DialogProperties
import com.example.axon.R
import com.example.axon.network.ApkDownloadResult
import com.example.axon.network.AppUpdater
import com.example.axon.network.DownloadFailure
import com.example.axon.network.PendingInstall
import com.example.axon.network.PendingInstallResult
import com.example.axon.network.UpdateCheckResult
import java.io.File
import kotlinx.coroutines.launch
import kotlin.math.cos
import kotlin.math.sin
private val BgGradStart  = Color(0xFF0F172A)
private val BgGradEnd    = Color(0xFF020617)
private val AccentGlow   = Color(0xFF1E3A8A)
private val CardColor    = Color(0xFF0F172A)
private val CardInner    = Color(0xFF1E293B)
private val BlueAccent   = Color(0xFF3B82F6)
private val VioletAccent = Color(0xFF8B5CF6)
private val TextPrimary  = Color(0xFFF9FAFB)
private val TextSecond   = Color(0xFF94A3B8)
private val BorderColor  = Color(0xFF334155)
private val px = floatArrayOf(0.15f, 0.85f, 0.25f, 0.75f, 0.5f, 0.35f, 0.65f, 0.2f, 0.8f, 0.9f)
private val py = floatArrayOf(0.2f, 0.15f, 0.75f, 0.8f, 0.65f, 0.45f, 0.35f, 0.55f, 0.6f, 0.4f)
private val pr = floatArrayOf(4f, 6f, 5f, 8f, 5f, 7f, 4f, 6f, 5f, 7f)
@Composable
fun StartScreen(
    onStart: () -> Unit,
    currentLanguage: String,
    onSelectLanguage: (String) -> Unit,
    updater: AppUpdater
) {
    var showGuide by remember { mutableStateOf(false) }
    var showLicense by remember { mutableStateOf(false) }
    // In-app updater state. The automatic check runs once per launch and
    // stays silent unless a newer version exists; the manual check (below)
    // always reports its outcome.
    val context = LocalContext.current
    val updateScope = rememberCoroutineScope()
    var updateState by remember { mutableStateOf<UpdateUiState>(UpdateUiState.Idle) }
    // Install whose outcome is not yet known: the system installer is in
    // front, or the user already came back from it. Kept outside
    // UpdateUiState so the dialog can be hidden while the installer
    // runs without losing track of the file.
    var pendingInstall by remember { mutableStateOf<PendingInstall?>(null) }
    // Set when the installer returns after a cancel: the next ON_RESUME
    // stays silent (a cancel is not an error), and only a later resume
    // offers the retry.
    var suppressResumeOfferOnce by remember { mutableStateOf(false) }

    /**
     * Learns the install outcome when the user returns from the system
     * installer. The installer result code is unreliable across OEMs
     * (many return CANCELED even on success, and on success the app
     * process is usually killed and replaced anyway), so it is never
     * trusted alone: the installed version is re-checked instead. The
     * APK is deleted only when the new version is confirmed installed.
     */
    val installLauncher = rememberLauncherForActivityResult(
        ActivityResultContracts.StartActivityForResult()
    ) { _ ->
        val pending = pendingInstall
        pendingInstall = null
        if (pending != null && updater.isVersionInstalled(pending.versionName)) {
            updater.deleteDownloadedApk()
            updater.clearPendingInstall()
            updateState = UpdateUiState.Idle
        } else {
            // Cancelled by the user, or the install did not complete:
            // dismiss silently — no error dialog. The APK and the
            // pending record are kept, so the next Start-screen resume
            // (or launch) offers a retry without another download.
            suppressResumeOfferOnce = true
            updateState = UpdateUiState.Idle
        }
    }

    /**
     * Fires the system installer for a downloaded APK. The permission is
     * re-checked at tap time, and the file is re-verified right before
     * firing, so a corrupt file is never handed to the installer (that
     * is what produces the system's "problem parsing the package"
     * error): instead the broken file is deleted and a fresh download
     * is offered. On a successful launch the APK is NOT deleted — the
     * launcher callback (and the resume check below) settle it once
     * the outcome is known.
     */
    fun fireInstall(file: File, versionName: String, downloadUrl: String, assetSize: Long) {
        // Re-checked at tap time: the user may just have come back from
        // the system settings where they granted (or refused) it.
        if (!updater.canInstallUnknownSources()) {
            updateState = UpdateUiState.NeedsPermission(
                file = file,
                versionName = versionName,
                downloadUrl = downloadUrl,
                assetSize = assetSize
            )
            return
        }
        if (!updater.verifyDownloadedApk(file, assetSize)) {
            updater.deleteDownloadedApk()
            updateState = UpdateUiState.DownloadFailed(
                downloadUrl = downloadUrl,
                reason = DownloadFailure.CORRUPT,
                versionName = versionName,
                assetSize = assetSize
            )
            return
        }
        val intent = updater.buildInstallIntent(context, file, assetSize)
        if (intent == null) {
            // FileProvider path mismatch or no installer: do not fire a
            // broken intent; offer a fresh download instead.
            updateState = UpdateUiState.DownloadFailed(
                downloadUrl = downloadUrl,
                reason = DownloadFailure.INSTALL_FAILED,
                versionName = versionName,
                assetSize = assetSize
            )
            return
        }
        updater.markInstallStarted(versionName, downloadUrl, assetSize)
        pendingInstall = PendingInstall(
            versionName = versionName,
            downloadUrl = downloadUrl,
            assetSize = assetSize,
            file = file
        )
        // Hide the dialog while the system installer is in front.
        updateState = UpdateUiState.Idle
        installLauncher.launch(intent)
    }

    LaunchedEffect(Unit) {
        // Settle any install started in a previous run: if the new
        // version is installed the APK is deleted silently; if not,
        // offer a retry with the kept file.
        when (val pending = updater.resolvePendingInstall()) {
            is PendingInstallResult.StillPending -> {
                updateState = UpdateUiState.InstallIncomplete(
                    file = pending.pending.file,
                    versionName = pending.pending.versionName,
                    downloadUrl = pending.pending.downloadUrl,
                    assetSize = pending.pending.assetSize
                )
            }
            PendingInstallResult.Installed, PendingInstallResult.None -> Unit
        }
        when (val result = updater.checkForUpdate()) {
            is UpdateCheckResult.UpdateAvailable -> {
                // Do not clobber the retry offer above with a fresh
                // offer for the same version.
                if (updateState is UpdateUiState.Idle) {
                    updateState = UpdateUiState.Available(
                        result.versionName,
                        result.downloadUrl,
                        result.assetSize
                    )
                }
            }
            // UpToDate / CheckFailed: silent on the automatic check.
            UpdateCheckResult.UpToDate, UpdateCheckResult.CheckFailed -> Unit
        }
    }

    // Re-check the installed version whenever the Start screen resumes:
    // this is what turns "install intent fired but install did not
    // complete" into a retry offer after the user comes back — including
    // the case where the installer returned without the launcher
    // callback running (e.g. the process was recreated while the
    // installer was in front).
    val lifecycleOwner = LocalLifecycleOwner.current
    DisposableEffect(lifecycleOwner) {
        val observer = LifecycleEventObserver { _, event ->
            if (event == Lifecycle.Event.ON_RESUME) {
                if (suppressResumeOfferOnce) {
                    suppressResumeOfferOnce = false
                    return@LifecycleEventObserver
                }
                if (updateState is UpdateUiState.Idle) {
                    when (val pending = updater.resolvePendingInstall()) {
                        is PendingInstallResult.StillPending -> {
                            updateState = UpdateUiState.InstallIncomplete(
                                file = pending.pending.file,
                                versionName = pending.pending.versionName,
                                downloadUrl = pending.pending.downloadUrl,
                                assetSize = pending.pending.assetSize
                            )
                        }
                        PendingInstallResult.Installed, PendingInstallResult.None -> Unit
                    }
                }
            }
        }
        lifecycleOwner.lifecycle.addObserver(observer)
        onDispose { lifecycleOwner.lifecycle.removeObserver(observer) }
    }

    fun runManualCheck() {
        if (updateState is UpdateUiState.Checking ||
            updateState is UpdateUiState.Downloading
        ) return
        updateState = UpdateUiState.Checking
        updateScope.launch {
            updateState = when (val result = updater.checkForUpdate(force = true)) {
                is UpdateCheckResult.UpdateAvailable ->
                    UpdateUiState.Available(result.versionName, result.downloadUrl, result.assetSize)
                UpdateCheckResult.UpToDate -> UpdateUiState.UpToDate
                UpdateCheckResult.CheckFailed -> UpdateUiState.CheckFailed
            }
        }
    }

    fun startDownload(downloadUrl: String, versionName: String, assetSize: Long) {
        updateState = UpdateUiState.Downloading(progress = -1)
        updateScope.launch {
            val result = updater.downloadApk(downloadUrl, assetSize) { percent ->
                // Dismissing the dialog cancels the download; never let a
                // late progress callback resurrect the dialog afterwards.
                if (updateState is UpdateUiState.Downloading) {
                    updateState = UpdateUiState.Downloading(progress = percent)
                }
            }
            // Dismissed mid-download (which cancels it): stay dismissed,
            // even if the download raced to completion — a completed
            // file that nobody asked to install is deleted, since no
            // install was started for it.
            if (updateState !is UpdateUiState.Downloading) {
                if (result is ApkDownloadResult.Success) result.file.delete()
                return@launch
            }
            when (result) {
                is ApkDownloadResult.Success -> {
                    fireInstall(result.file, versionName, downloadUrl, assetSize)
                }
                is ApkDownloadResult.Failed -> {
                    updateState = UpdateUiState.DownloadFailed(
                        downloadUrl = downloadUrl,
                        reason = result.reason,
                        versionName = versionName,
                        assetSize = assetSize
                    )
                }
            }
        }
    }
    val fadeIn   = remember { Animatable(0f) }
    val slideUp  = remember { Animatable(30f) }
    val btnScale = remember { Animatable(0.9f) }
    LaunchedEffect(Unit) {
        fadeIn.animateTo(1f, tween(800, easing = EaseOutCubic))
        slideUp.animateTo(0f, tween(800, easing = EaseOutCubic))
        btnScale.animateTo(1f, tween(600, easing = EaseOutCubic))
    }
    val transition = rememberInfiniteTransition(label = "pulse")
    val timeValue by transition.animateFloat(
        initialValue = 0f,
        targetValue = 2f * Math.PI.toFloat(),
        animationSpec = infiniteRepeatable(
            animation = tween(12000, easing = LinearEasing),
            repeatMode = RepeatMode.Restart
        ),
        label = "time"
    )
    val pulseA by transition.animateFloat(
        initialValue = 0.3f,
        targetValue = 0.8f,
        animationSpec = infiniteRepeatable(tween(2500), RepeatMode.Reverse),
        label = "pa"
    )
    val pulseR by transition.animateFloat(
        initialValue = 0.85f,
        targetValue = 1.15f,
        animationSpec = infiniteRepeatable(tween(2500), RepeatMode.Reverse),
        label = "pr"
    )
    val floatOffset by transition.animateFloat(
        initialValue = -5f,
        targetValue = 5f,
        animationSpec = infiniteRepeatable(tween(3000), RepeatMode.Reverse),
        label = "float"
    )
    Box(
        modifier = Modifier
            .fillMaxSize()
            .background(
                Brush.verticalGradient(
                    colors = listOf(BgGradStart, BgGradEnd)
                )
            )
    ) {
        Canvas(modifier = Modifier.fillMaxSize()) {
            drawCircle(
                color = AccentGlow.copy(alpha = 0.35f),
                radius = size.minDimension * 0.7f,
                center = Offset(size.width / 2f, size.height * 0.3f)
            )
            val twelvePx = 12.dp.toPx()
            for (i in px.indices) {
                val dx = sin(timeValue + i * 1.5f) * twelvePx
                val dy = cos(timeValue + i * 2.0f) * twelvePx
                val x = size.width * px[i] + dx
                val y = size.height * py[i] + dy
                drawCircle(
                    color = BlueAccent.copy(alpha = 0.12f),
                    radius = pr[i].dp.toPx(),
                    center = Offset(x, y)
                )
            }
        }
        Column(
            modifier = Modifier
                .fillMaxSize()
                .padding(horizontal = 24.dp)
                .alpha(fadeIn.value)
                .graphicsLayer { translationY = slideUp.value },
            horizontalAlignment = Alignment.CenterHorizontally
        ) {
            Spacer(modifier = Modifier.height(48.dp))
            // Language selector: all three languages are always shown and
            // directly selectable. The old selector only toggled ES <-> EN,
            // so once the user left Chinese they could never go back (#11).
            Row(
                modifier = Modifier
                    .clip(RoundedCornerShape(24.dp))
                    .background(Color.White.copy(alpha = 0.06f))
                    .border(1.dp, Color.White.copy(alpha = 0.12f), RoundedCornerShape(24.dp))
                    .padding(horizontal = 8.dp, vertical = 6.dp),
                verticalAlignment = Alignment.CenterVertically
            ) {
                listOf(
                    "en" to "🇺🇸 EN",
                    "es" to "🇪🇸 ES",
                    "zh" to "🇨🇳 中文"
                ).forEachIndexed { index, (code, label) ->
                    if (index > 0) {
                        Text(
                            text = "|",
                            color = Color.White.copy(alpha = 0.2f),
                            fontSize = 12.sp,
                            modifier = Modifier.padding(horizontal = 4.dp)
                        )
                    }
                    Text(
                        text = label,
                        color = if (currentLanguage == code) TextPrimary else TextSecond,
                        fontSize = 12.sp,
                        fontWeight = if (currentLanguage == code) FontWeight.Bold else FontWeight.Normal,
                        modifier = Modifier
                            .clip(RoundedCornerShape(16.dp))
                            .clickable { onSelectLanguage(code) }
                            .padding(horizontal = 10.dp, vertical = 6.dp)
                    )
                }
            }
            Spacer(modifier = Modifier.weight(1f))
            Box(
                modifier = Modifier
                    .size(150.dp)
                    .graphicsLayer { translationY = floatOffset },
                contentAlignment = Alignment.Center
            ) {
                Canvas(modifier = Modifier.fillMaxSize()) {
                    val cx = size.width / 2f
                    val cy = size.height / 2f
                    val r = size.minDimension / 2f
                    drawCircle(
                        color = BlueAccent.copy(alpha = pulseA * 0.10f),
                        radius = r * pulseR,
                        center = Offset(cx, cy)
                    )
                    drawCircle(
                        brush = Brush.radialGradient(
                            colors = listOf(BlueAccent.copy(alpha = 0.25f), Color.Transparent),
                            center = Offset(cx, cy),
                            radius = r * 0.85f
                        ),
                        radius = r * 0.85f,
                        center = Offset(cx, cy)
                    )
                    drawCircle(
                        color = BlueAccent.copy(alpha = 0.08f),
                        radius = r * 0.7f,
                        center = Offset(cx, cy),
                        style = Stroke(width = 2.dp.toPx())
                    )
                    val nodes = 6
                    for (i in 0 until nodes) {
                        val angle = (i * 2 * Math.PI / nodes).toFloat()
                        val nx = cx + cos(angle) * r * 0.65f
                        val ny = cy + sin(angle) * r * 0.65f
                        drawCircle(
                            color = BlueAccent.copy(alpha = 0.4f),
                            radius = 3.dp.toPx(),
                            center = Offset(nx, ny)
                        )
                        drawLine(
                            color = BlueAccent.copy(alpha = 0.15f),
                            start = Offset(cx, cy),
                            end = Offset(nx, ny),
                            strokeWidth = 1.dp.toPx()
                        )
                    }
                    drawCircle(
                        color = BlueAccent,
                        radius = 5.dp.toPx(),
                        center = Offset(cx, cy)
                    )
                    val iconSize = r * 0.45f
                    val left  = cx - iconSize * 0.6f
                    val right = cx + iconSize * 0.6f
                    val top   = cy - iconSize * 0.7f
                    val base  = cy + iconSize * 0.5f
                    val mid   = cy - iconSize * 0.05f
                    val aPath = Path().apply {
                        moveTo(cx, top)
                        lineTo(right, base)
                        moveTo(left, base)
                        lineTo(cx, top)
                        moveTo(left + (right - left) * 0.22f, mid)
                        lineTo(right - (right - left) * 0.22f, mid)
                    }
                    drawPath(aPath, color = TextPrimary, style = Stroke(width = 4.dp.toPx(), cap = StrokeCap.Round))
                    val curW = iconSize * 0.32f
                    val curH = iconSize * 0.4f
                    val curX = cx + iconSize * 0.18f
                    val curY = cy + iconSize * 0.05f
                    val cursorPath = Path().apply {
                        moveTo(curX, curY)
                        lineTo(curX, curY + curH)
                        lineTo(curX + curW * 0.38f, curY + curH * 0.72f)
                        lineTo(curX + curW * 0.65f, curY + curH)
                        lineTo(curX + curW * 0.78f, curY + curH * 0.9f)
                        lineTo(curX + curW * 0.5f, curY + curH * 0.62f)
                        lineTo(curX + curW, curY + curH * 0.62f)
                        close()
                    }
                    drawPath(cursorPath, color = VioletAccent)
                }
            }
            Spacer(modifier = Modifier.height(24.dp))
            Text(
                text = "AXON",
                color = TextPrimary,
                fontSize = 44.sp,
                fontWeight = FontWeight.ExtraBold,
                letterSpacing = 2.sp
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = stringResource(id = R.string.remote_desktop_control),
                color = TextSecond,
                fontSize = 15.sp,
                fontWeight = FontWeight.Medium,
                textAlign = TextAlign.Center
            )
            Spacer(modifier = Modifier.weight(1f))
            Card(
                modifier = Modifier
                    .fillMaxWidth()
                    .border(1.dp, BorderColor, RoundedCornerShape(20.dp)),
                shape = RoundedCornerShape(20.dp),
                colors = CardDefaults.cardColors(containerColor = CardColor.copy(alpha = 0.85f))
            ) {
                Column(
                    modifier = Modifier.padding(24.dp),
                    horizontalAlignment = Alignment.CenterHorizontally
                ) {
                    Text(
                        text = stringResource(id = R.string.remote_mouse),
                        color = TextPrimary,
                        fontSize = 16.sp,
                        fontWeight = FontWeight.Bold
                    )
                    Spacer(modifier = Modifier.height(6.dp))
                    Text(
                        text = stringResource(id = R.string.connect_instructions),
                        color = TextSecond,
                        fontSize = 13.sp,
                        textAlign = TextAlign.Center,
                        lineHeight = 18.sp
                    )
                }
            }
            Spacer(modifier = Modifier.height(24.dp))
            Button(
                onClick = onStart,
                modifier = Modifier
                    .fillMaxWidth()
                    .height(56.dp)
                    .graphicsLayer {
                        scaleX = btnScale.value
                        scaleY = btnScale.value
                    },
                shape = RoundedCornerShape(16.dp),
                colors = ButtonDefaults.buttonColors(
                    containerColor = Color.Transparent,
                    contentColor = Color.White
                ),
                contentPadding = PaddingValues()
            ) {
                Box(
                    modifier = Modifier
                        .fillMaxSize()
                        .background(
                            Brush.horizontalGradient(
                                colors = listOf(BlueAccent, VioletAccent)
                            )
                        ),
                    contentAlignment = Alignment.Center
                ) {
                    Text(
                        text = stringResource(id = R.string.connect_to_desktop),
                        fontSize = 16.sp,
                        fontWeight = FontWeight.Bold,
                        letterSpacing = 0.5.sp
                    )
                }
            }
            Spacer(modifier = Modifier.height(16.dp))
            Row(
                modifier = Modifier.fillMaxWidth(),
                horizontalArrangement = Arrangement.Center
            ) {
                TextButton(onClick = { showGuide = true }) {
                    Text(
                        text = stringResource(id = R.string.button_guide),
                        color = BlueAccent,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
                Spacer(modifier = Modifier.width(12.dp))
                TextButton(onClick = { showLicense = true }) {
                    Text(
                        text = stringResource(id = R.string.button_license),
                        color = BlueAccent,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
                Spacer(modifier = Modifier.width(12.dp))
                TextButton(
                    onClick = { runManualCheck() },
                    enabled = updateState !is UpdateUiState.Checking &&
                        updateState !is UpdateUiState.Downloading
                ) {
                    Text(
                        text = stringResource(
                            id = if (updateState is UpdateUiState.Checking) {
                                R.string.update_checking
                            } else {
                                R.string.update_check
                            }
                        ),
                        color = BlueAccent,
                        fontSize = 14.sp,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
            Spacer(modifier = Modifier.height(24.dp))
            Text(
                text = "Alenia Studios · GPL v3",
                color = TextSecond.copy(alpha = 0.5f),
                fontSize = 11.sp,
                fontWeight = FontWeight.Medium
            )
            Spacer(modifier = Modifier.height(16.dp))
        }
    }
    if (showGuide) {
        ConnectionGuideDialog(onDismiss = { showGuide = false })
    }
    if (showLicense) {
        LicenseDialog(onDismiss = { showLicense = false })
    }
    when (val state = updateState) {
        UpdateUiState.Idle, UpdateUiState.Checking -> Unit
        is UpdateUiState.Available,
        is UpdateUiState.Downloading,
        is UpdateUiState.DownloadFailed,
        is UpdateUiState.NeedsPermission,
        is UpdateUiState.InstallIncomplete -> {
            UpdateDialog(
                state = state,
                onDownload = { url ->
                    // The version/size travel with the state so a retry
                    // re-download verifies against the same expected size.
                    val (version, size) = when (state) {
                        is UpdateUiState.Available -> state.versionName to state.assetSize
                        is UpdateUiState.DownloadFailed -> state.versionName to state.assetSize
                        is UpdateUiState.NeedsPermission -> state.versionName to state.assetSize
                        is UpdateUiState.InstallIncomplete -> state.versionName to state.assetSize
                        else -> "" to -1L
                    }
                    startDownload(url, version, size)
                },
                onOpenSettings = { updater.openUnknownSourcesSettings(context) },
                onInstall = { file ->
                    // Re-checked inside fireInstall: the user may just
                    // have come back from the system settings where they
                    // granted the permission.
                    val (version, url, size) = when (state) {
                        is UpdateUiState.NeedsPermission ->
                            Triple(state.versionName, state.downloadUrl, state.assetSize)
                        is UpdateUiState.InstallIncomplete ->
                            Triple(state.versionName, state.downloadUrl, state.assetSize)
                        else -> Triple("", "", -1L)
                    }
                    fireInstall(file, version, url, size)
                },
                onDismiss = {
                    if (state is UpdateUiState.Downloading) updater.cancelDownload()
                    // Dismissing InstallIncomplete / NeedsPermission keeps
                    // the downloaded APK and the pending record: the
                    // resume check offers the retry again later.
                    updateState = UpdateUiState.Idle
                }
            )
        }
        UpdateUiState.UpToDate -> {
            UpdateStatusDialog(
                title = stringResource(id = R.string.update_up_to_date),
                message = stringResource(id = R.string.update_up_to_date_message),
                onDismiss = { updateState = UpdateUiState.Idle }
            )
        }
        UpdateUiState.CheckFailed -> {
            UpdateStatusDialog(
                title = stringResource(id = R.string.update_check_failed),
                message = stringResource(id = R.string.update_check_failed_message),
                onDismiss = { updateState = UpdateUiState.Idle }
            )
        }
    }
}

/**
 * State of the in-app updater UI on the Start screen. One value drives
 * the whole dialog, so the dialog can never show a stale combination
 * (e.g. progress without a download).
 */
private sealed interface UpdateUiState {
    data object Idle : UpdateUiState
    data object Checking : UpdateUiState
    data class Available(
        val versionName: String,
        val downloadUrl: String,
        val assetSize: Long = -1
    ) : UpdateUiState
    /** [progress] is 0-100, or -1 while the download size is unknown. */
    data class Downloading(val progress: Int) : UpdateUiState
    data class DownloadFailed(
        val downloadUrl: String,
        val reason: DownloadFailure = DownloadFailure.NETWORK,
        val versionName: String = "",
        val assetSize: Long = -1
    ) : UpdateUiState
    /** APK downloaded, but the install-unknown-apps grant is missing. */
    data class NeedsPermission(
        val file: File,
        val versionName: String = "",
        val downloadUrl: String = "",
        val assetSize: Long = -1
    ) : UpdateUiState
    /**
     * Install was started but the old version is still installed when
     * the user came back: the APK was kept, so install can be retried
     * without downloading again.
     */
    data class InstallIncomplete(
        val file: File,
        val versionName: String,
        val downloadUrl: String,
        val assetSize: Long = -1
    ) : UpdateUiState
    data object UpToDate : UpdateUiState
    data object CheckFailed : UpdateUiState
}

/**
 * The update dialog. In-app only: it never opens a browser or GitHub.
 * Its content follows [UpdateUiState]: offer / progress / failure with
 * retry / install-permission guidance.
 */
@Composable
private fun UpdateDialog(
    state: UpdateUiState,
    onDownload: (String) -> Unit,
    onOpenSettings: () -> Unit,
    onInstall: (File) -> Unit,
    onDismiss: () -> Unit
) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth(0.9f)
                .border(1.dp, BorderColor, RoundedCornerShape(24.dp)),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = BgGradStart)
        ) {
            Column(modifier = Modifier.padding(24.dp)) {
                when (state) {
                    is UpdateUiState.Available -> {
                        Text(
                            text = stringResource(id = R.string.update_available),
                            color = TextPrimary,
                            fontSize = 20.sp,
                            fontWeight = FontWeight.ExtraBold
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Text(
                            text = stringResource(
                                id = R.string.update_available_message,
                                state.versionName
                            ),
                            color = TextSecond,
                            fontSize = 14.sp,
                            lineHeight = 20.sp
                        )
                        Spacer(modifier = Modifier.height(20.dp))
                        Button(
                            onClick = { onDownload(state.downloadUrl) },
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = BlueAccent)
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_download),
                                color = TextPrimary,
                                fontWeight = FontWeight.Bold
                            )
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        TextButton(
                            onClick = onDismiss,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_later),
                                color = TextSecond,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }
                    is UpdateUiState.Downloading -> {
                        Text(
                            text = if (state.progress >= 0) {
                                stringResource(
                                    id = R.string.update_download_progress,
                                    state.progress
                                )
                            } else {
                                stringResource(id = R.string.update_downloading)
                            },
                            color = TextPrimary,
                            fontSize = 16.sp,
                            fontWeight = FontWeight.Bold
                        )
                        Spacer(modifier = Modifier.height(16.dp))
                        if (state.progress >= 0) {
                            LinearProgressIndicator(
                                progress = { state.progress / 100f },
                                modifier = Modifier.fillMaxWidth(),
                                color = BlueAccent,
                                trackColor = CardInner
                            )
                        } else {
                            LinearProgressIndicator(
                                modifier = Modifier.fillMaxWidth(),
                                color = BlueAccent,
                                trackColor = CardInner
                            )
                        }
                        Spacer(modifier = Modifier.height(20.dp))
                        TextButton(
                            onClick = onDismiss,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_later),
                                color = TextSecond,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }
                    is UpdateUiState.DownloadFailed -> {
                        Text(
                            text = stringResource(id = R.string.update_download_failed),
                            color = TextPrimary,
                            fontSize = 20.sp,
                            fontWeight = FontWeight.ExtraBold
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Text(
                            // The reason decides the message: a truncated
                            // or corrupt download says so plainly, an
                            // unverifiable one says it could not be
                            // checked, and a failed install intent says
                            // the file could not be opened — each with a
                            // retry that downloads again.
                            text = stringResource(
                                id = when (state.reason) {
                                    DownloadFailure.SIZE_MISMATCH,
                                    DownloadFailure.CORRUPT ->
                                        R.string.update_download_failed_size_message
                                    DownloadFailure.UNVERIFIABLE ->
                                        R.string.update_download_failed_unverifiable_message
                                    DownloadFailure.INSTALL_FAILED ->
                                        R.string.update_install_failed_message
                                    DownloadFailure.NETWORK ->
                                        R.string.update_download_failed_message
                                }
                            ),
                            color = TextSecond,
                            fontSize = 14.sp,
                            lineHeight = 20.sp
                        )
                        Spacer(modifier = Modifier.height(20.dp))
                        Button(
                            onClick = { onDownload(state.downloadUrl) },
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = BlueAccent)
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_retry),
                                color = TextPrimary,
                                fontWeight = FontWeight.Bold
                            )
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        TextButton(
                            onClick = onDismiss,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_later),
                                color = TextSecond,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }
                    is UpdateUiState.NeedsPermission -> {
                        Text(
                            text = stringResource(id = R.string.update_install_permission_title),
                            color = TextPrimary,
                            fontSize = 20.sp,
                            fontWeight = FontWeight.ExtraBold
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Text(
                            text = stringResource(id = R.string.update_install_permission_message),
                            color = TextSecond,
                            fontSize = 14.sp,
                            lineHeight = 20.sp
                        )
                        Spacer(modifier = Modifier.height(20.dp))
                        Button(
                            onClick = onOpenSettings,
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = BlueAccent)
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_open_settings),
                                color = TextPrimary,
                                fontWeight = FontWeight.Bold
                            )
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(
                            onClick = { onInstall(state.file) },
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = CardInner)
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_install),
                                color = TextPrimary,
                                fontWeight = FontWeight.Bold
                            )
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        TextButton(
                            onClick = onDismiss,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_later),
                                color = TextSecond,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }
                    is UpdateUiState.InstallIncomplete -> {
                        Text(
                            text = stringResource(id = R.string.update_install_incomplete_title),
                            color = TextPrimary,
                            fontSize = 20.sp,
                            fontWeight = FontWeight.ExtraBold
                        )
                        Spacer(modifier = Modifier.height(12.dp))
                        Text(
                            text = stringResource(
                                id = R.string.update_install_incomplete_message,
                                state.versionName
                            ),
                            color = TextSecond,
                            fontSize = 14.sp,
                            lineHeight = 20.sp
                        )
                        Spacer(modifier = Modifier.height(20.dp))
                        Button(
                            onClick = { onInstall(state.file) },
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = BlueAccent)
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_install_again),
                                color = TextPrimary,
                                fontWeight = FontWeight.Bold
                            )
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        Button(
                            onClick = { onDownload(state.downloadUrl) },
                            modifier = Modifier.fillMaxWidth(),
                            shape = RoundedCornerShape(12.dp),
                            colors = ButtonDefaults.buttonColors(containerColor = CardInner)
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_retry),
                                color = TextPrimary,
                                fontWeight = FontWeight.Bold
                            )
                        }
                        Spacer(modifier = Modifier.height(8.dp))
                        TextButton(
                            onClick = onDismiss,
                            modifier = Modifier.fillMaxWidth()
                        ) {
                            Text(
                                text = stringResource(id = R.string.update_later),
                                color = TextSecond,
                                fontWeight = FontWeight.Bold
                            )
                        }
                    }
                    UpdateUiState.Idle, UpdateUiState.Checking,
                    UpdateUiState.UpToDate, UpdateUiState.CheckFailed -> Unit
                }
            }
        }
    }
}

/** Small one-message dialog for the manual check outcome. */
@Composable
private fun UpdateStatusDialog(
    title: String,
    message: String,
    onDismiss: () -> Unit
) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth(0.9f)
                .border(1.dp, BorderColor, RoundedCornerShape(24.dp)),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = BgGradStart)
        ) {
            Column(modifier = Modifier.padding(24.dp)) {
                Text(
                    text = title,
                    color = TextPrimary,
                    fontSize = 20.sp,
                    fontWeight = FontWeight.ExtraBold
                )
                Spacer(modifier = Modifier.height(12.dp))
                Text(
                    text = message,
                    color = TextSecond,
                    fontSize = 14.sp,
                    lineHeight = 20.sp
                )
                Spacer(modifier = Modifier.height(20.dp))
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = CardInner)
                ) {
                    Text(
                        text = stringResource(id = R.string.close_action),
                        color = TextPrimary,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}
@Composable
fun ConnectionGuideDialog(onDismiss: () -> Unit) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth(0.9f)
                .fillMaxHeight(0.8f)
                .border(1.dp, BorderColor, RoundedCornerShape(24.dp)),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = BgGradStart)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(24.dp)
            ) {
                Text(
                    text = stringResource(id = R.string.guide_title),
                    color = TextPrimary,
                    fontSize = 22.sp,
                    fontWeight = FontWeight.ExtraBold,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth()
                )
                Spacer(modifier = Modifier.height(20.dp))
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .verticalScroll(rememberScrollState())
                ) {
                    GuideSection(
                        title = stringResource(id = R.string.guide_wifi_title),
                        description = stringResource(id = R.string.guide_wifi_step)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    GuideSection(
                        title = stringResource(id = R.string.guide_usb_title),
                        description = stringResource(id = R.string.guide_usb_step)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    GuideSection(
                        title = stringResource(id = R.string.guide_bt_title),
                        description = stringResource(id = R.string.guide_bt_step)
                    )
                }
                Spacer(modifier = Modifier.height(20.dp))
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = CardInner)
                ) {
                    Text(
                        text = stringResource(id = R.string.close_action),
                        color = TextPrimary,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}
@Composable
fun GuideSection(title: String, description: String) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(1.dp, BorderColor.copy(alpha = 0.6f), RoundedCornerShape(16.dp)),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = CardInner)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = title,
                color = BlueAccent,
                fontSize = 15.sp,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = description,
                color = TextSecond,
                fontSize = 13.sp,
                lineHeight = 18.sp
            )
        }
    }
}
@Composable
fun LicenseDialog(onDismiss: () -> Unit) {
    Dialog(
        onDismissRequest = onDismiss,
        properties = DialogProperties(usePlatformDefaultWidth = false)
    ) {
        Card(
            modifier = Modifier
                .fillMaxWidth(0.9f)
                .fillMaxHeight(0.8f)
                .border(1.dp, BorderColor, RoundedCornerShape(24.dp)),
            shape = RoundedCornerShape(24.dp),
            colors = CardDefaults.cardColors(containerColor = BgGradStart)
        ) {
            Column(
                modifier = Modifier
                    .fillMaxSize()
                    .padding(24.dp)
            ) {
                Text(
                    text = stringResource(id = R.string.license_title),
                    color = TextPrimary,
                    fontSize = 22.sp,
                    fontWeight = FontWeight.ExtraBold,
                    textAlign = TextAlign.Center,
                    modifier = Modifier.fillMaxWidth()
                )
                Spacer(modifier = Modifier.height(20.dp))
                Column(
                    modifier = Modifier
                        .weight(1f)
                        .verticalScroll(rememberScrollState())
                ) {
                    LicenseSection(
                        title = stringResource(id = R.string.license_software_title),
                        description = stringResource(id = R.string.license_software_desc)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    LicenseSection(
                        title = stringResource(id = R.string.license_assets_title),
                        description = stringResource(id = R.string.license_assets_desc)
                    )
                    Spacer(modifier = Modifier.height(16.dp))
                    LicenseSection(
                        title = stringResource(id = R.string.license_contact_title),
                        description = stringResource(id = R.string.license_contact_desc)
                    )
                }
                Spacer(modifier = Modifier.height(20.dp))
                Button(
                    onClick = onDismiss,
                    modifier = Modifier.fillMaxWidth(),
                    shape = RoundedCornerShape(12.dp),
                    colors = ButtonDefaults.buttonColors(containerColor = CardInner)
                ) {
                    Text(
                        text = stringResource(id = R.string.close_action),
                        color = TextPrimary,
                        fontWeight = FontWeight.Bold
                    )
                }
            }
        }
    }
}
@Composable
fun LicenseSection(title: String, description: String) {
    Card(
        modifier = Modifier
            .fillMaxWidth()
            .border(1.dp, BorderColor.copy(alpha = 0.6f), RoundedCornerShape(16.dp)),
        shape = RoundedCornerShape(16.dp),
        colors = CardDefaults.cardColors(containerColor = CardInner)
    ) {
        Column(modifier = Modifier.padding(16.dp)) {
            Text(
                text = title,
                color = VioletAccent,
                fontSize = 15.sp,
                fontWeight = FontWeight.Bold
            )
            Spacer(modifier = Modifier.height(8.dp))
            Text(
                text = description,
                color = TextSecond,
                fontSize = 13.sp,
                lineHeight = 18.sp
            )
        }
    }
}
