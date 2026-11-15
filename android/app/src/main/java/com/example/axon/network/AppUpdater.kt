package com.example.axon.network

import android.content.Context
import android.content.Intent
import android.content.pm.PackageManager
import android.net.Uri
import android.os.Build
import android.provider.Settings
import androidx.core.content.FileProvider
import kotlinx.coroutines.Dispatchers
import kotlinx.coroutines.withContext
import okhttp3.OkHttpClient
import okhttp3.Request
import org.json.JSONObject
import java.io.File
import java.io.FileOutputStream
import java.util.concurrent.TimeUnit

/**
 * In-app self-updater for the Axon Android app.
 *
 * Flow (never opens a browser or GitHub):
 *  1. [checkForUpdate] GETs the latest release from the GitHub API and looks
 *     for the `axon-android-debug.apk` asset, recording its `size`.
 *  2. If the release tag is newer than the installed version, the Start
 *     screen shows an in-app dialog with a Download button.
 *  3. [downloadApk] fetches the APK into the stable filesDir/updates/
 *     dir with progress, then verifies it before anything else happens:
 *     bytes written must equal the GitHub asset size (or the
 *     Content-Length fallback); if neither size is known the download is
 *     rejected as unverifiable. The file is flushed and fsynced before
 *     it is handed anywhere.
 *  4. [buildInstallIntent] builds an ACTION_INSTALL_PACKAGE intent for
 *     the verified file through a FileProvider URI; the Start screen
 *     fires it through an ActivityResultLauncher so the outcome is
 *     learned when the user returns. The APK is deleted only once the
 *     new version is confirmed installed (or a newer download
 *     supersedes it) — never before the install result is known.
 *
 * All network failures are silent ([UpdateCheckResult.CheckFailed] /
 * [ApkDownloadResult.Failed]) — an update check must never crash or
 * block startup. The check runs at most once per launch unless forced,
 * so the unauthenticated GitHub API rate limit (60 requests/hour per IP)
 * is never a concern.
 */
class AppUpdater(context: Context) {

    private val appContext: Context = context.applicationContext

    private val client: OkHttpClient = OkHttpClient.Builder()
        .connectTimeout(CHECK_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .readTimeout(CHECK_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .writeTimeout(CHECK_TIMEOUT_SECONDS, TimeUnit.SECONDS)
        .build()

    /** Result of the last automatic (non-forced) check, if one already ran. */
    private var lastResult: UpdateCheckResult? = null

    /** The in-flight download call, so the UI can cancel it on dismiss. */
    @Volatile
    private var downloadCall: okhttp3.Call? = null

    /**
     * Checks GitHub for a newer release.
     *
     * @param force when false (automatic check on launch) a cached result
     *  from this launch is returned instead of hitting the API again.
     *  Pass true for the manual "Check for updates" action.
     */
    suspend fun checkForUpdate(force: Boolean = false): UpdateCheckResult {
        lastResult?.let { cached ->
            if (!force) return cached
        }
        val result = runCheck()
        lastResult = result
        return result
    }

    private suspend fun runCheck(): UpdateCheckResult = withContext(Dispatchers.IO) {
        try {
            val request = Request.Builder()
                .url(RELEASES_URL)
                .header("Accept", "application/vnd.github+json")
                .header("User-Agent", USER_AGENT)
                .build()
            client.newCall(request).execute().use { response ->
                if (!response.isSuccessful) return@withContext UpdateCheckResult.CheckFailed
                val body = response.body?.string() ?: return@withContext UpdateCheckResult.CheckFailed
                parseRelease(body)
            }
        } catch (e: Exception) {
            // Offline, timeout, bad JSON, rate-limited (403): silent failure.
            UpdateCheckResult.CheckFailed
        }
    }

    /** Parses a `GET /repos/.../releases/latest` body. Pure logic, no I/O. */
    internal fun parseRelease(body: String): UpdateCheckResult {
        return try {
            val json = JSONObject(body)
            val tag = json.optString("tag_name", "").trim()
            if (tag.isEmpty()) return UpdateCheckResult.CheckFailed
            val assets = json.optJSONArray("assets") ?: return UpdateCheckResult.CheckFailed
            var downloadUrl: String? = null
            var assetSize = -1L
            for (i in 0 until assets.length()) {
                val asset = assets.optJSONObject(i) ?: continue
                if (asset.optString("name") == APK_ASSET_NAME) {
                    downloadUrl = asset.optString("browser_download_url", "")
                    // GitHub asset size in bytes; -1 when absent. This is
                    // the size the download is verified against — it is
                    // known before the download starts, unlike
                    // Content-Length, which depends on the download
                    // response (and is -1 for chunked responses).
                    assetSize = asset.optLong("size", -1L)
                    break
                }
            }
            if (downloadUrl.isNullOrBlank()) return UpdateCheckResult.CheckFailed
            val installed = installedVersion()
            if (isNewerVersion(tag, installed.versionName)) {
                UpdateCheckResult.UpdateAvailable(
                    versionName = tag.trimStart('v', 'V'),
                    downloadUrl = downloadUrl,
                    assetSize = assetSize
                )
            } else {
                UpdateCheckResult.UpToDate
            }
        } catch (e: Exception) {
            UpdateCheckResult.CheckFailed
        }
    }

    /** Installed version, from PackageManager (buildConfig is disabled). */
    fun installedVersion(): InstalledVersion {
        val packageManager = appContext.packageManager
        @Suppress("DEPRECATION")
        val info = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.TIRAMISU) {
            packageManager.getPackageInfo(
                appContext.packageName,
                PackageManager.PackageInfoFlags.of(0)
            )
        } else {
            packageManager.getPackageInfo(appContext.packageName, 0)
        }
        val code = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.P) {
            info.longVersionCode
        } else {
            @Suppress("DEPRECATION")
            info.versionCode.toLong()
        }
        return InstalledVersion(versionName = info.versionName ?: "0", versionCode = code)
    }

    /**
     * Stable home for the downloaded APK: filesDir/updates/. The old
     * location (cacheDir) could be cleaned by the system mid-install,
     * leaving the installer with a vanished file; filesDir is never
     * cleaned by the system. The FileProvider serves exactly this
     * directory (see res/xml/file_paths.xml, <files-path name="updates"
     * path="updates/" />); keep the two in sync.
     */
    fun updatesDir(): File =
        File(appContext.filesDir, UPDATES_DIR_NAME).apply { if (!exists()) mkdirs() }

    /** The downloaded APK file, in the stable location. */
    fun downloadedApkFile(): File = File(updatesDir(), APK_FILE_NAME)

    /**
     * Deletes the downloaded APK from the stable location and from the
     * legacy cache location. Only call this once the new version is
     * confirmed installed, when a newer download supersedes it, or when
     * the file itself failed verification.
     */
    fun deleteDownloadedApk() {
        downloadedApkFile().delete()
        File(appContext.cacheDir, APK_FILE_NAME).delete()
    }

    /**
     * True when [file] looks like a complete APK: it exists, is
     * non-empty, matches [expectedSize] when one is known, and starts
     * with the ZIP "PK" magic every APK carries. A cheap pre-flight
     * guard against truncated downloads and HTML error pages served in
     * place of the APK; the real package parse is still done by the
     * system installer. Call this again right before firing the install
     * intent — the file may have changed since the download finished.
     */
    fun verifyDownloadedApk(file: File, expectedSize: Long = -1): Boolean {
        if (!file.isFile || file.length() <= 0) return false
        if (expectedSize > 0 && file.length() != expectedSize) return false
        return try {
            file.inputStream().use { input ->
                val magic = ByteArray(2)
                input.read(magic) == 2 &&
                    magic[0] == 0x50.toByte() && magic[1] == 0x4B.toByte()
            }
        } catch (e: Exception) {
            false
        }
    }

    /**
     * Downloads the APK into the stable filesDir/updates/ dir, reporting
     * 0-100 progress (-1 while the size is still unknown), and verifies
     * it fully before returning success. The size to verify against is
     * the GitHub asset size passed in as [expectedSize] when known,
     * falling back to the download response's Content-Length. If neither
     * is known the download is rejected as unverifiable rather than
     * installed blind. The file is flushed and fsynced before the
     * stream closes, so the bytes are durable on disk before any
     * install intent is built from them.
     *
     * Returns [ApkDownloadResult.Success] with the verified file, or
     * [ApkDownloadResult.Failed] with the reason. A partial, truncated
     * or unverifiable file is always deleted on failure/cancel, so a
     * broken file is never left behind to be installed later.
     */
    suspend fun downloadApk(
        downloadUrl: String,
        expectedSize: Long = -1,
        onProgress: (Int) -> Unit
    ): ApkDownloadResult = withContext(Dispatchers.IO) {
        val target = downloadedApkFile()
        try {
            if (target.exists()) target.delete()
            // Legacy location from the first version: remove it so a
            // stale copy can never be picked up by mistake.
            File(appContext.cacheDir, APK_FILE_NAME).delete()
            val request = Request.Builder()
                .url(downloadUrl)
                .header("User-Agent", USER_AGENT)
                .build()
            val call = client.newCall(request)
            downloadCall = call
            call.execute().use { response ->
                if (!response.isSuccessful) {
                    return@withContext ApkDownloadResult.Failed(DownloadFailure.NETWORK)
                }
                val body = response.body
                    ?: return@withContext ApkDownloadResult.Failed(DownloadFailure.NETWORK)
                val headerSize = body.contentLength() // -1 when the server hides it
                val sizeToVerify = if (expectedSize > 0) expectedSize else headerSize
                if (sizeToVerify <= 0) {
                    // No trustworthy size from either source: refuse to
                    // produce a file that could never be verified.
                    return@withContext ApkDownloadResult.Failed(DownloadFailure.UNVERIFIABLE)
                }
                var downloaded = 0L
                var lastReported = -1
                body.byteStream().use { input ->
                    FileOutputStream(target).use { output ->
                        val buffer = ByteArray(DOWNLOAD_BUFFER_BYTES)
                        var read: Int
                        while (input.read(buffer).also { read = it } != -1) {
                            output.write(buffer, 0, read)
                            downloaded += read
                            val percent =
                                ((downloaded * 100) / sizeToVerify).toInt().coerceIn(0, 100)
                            if (percent != lastReported) {
                                lastReported = percent
                                withContext(Dispatchers.Main) { onProgress(percent) }
                            }
                        }
                        // Flush + fsync BEFORE the stream closes and
                        // before any install intent is built: the bytes
                        // must be durable on disk, not just in a buffer.
                        output.flush()
                        output.fd.sync()
                    }
                }
                // Guard against a truncated download the server called done.
                if (target.length() != sizeToVerify) {
                    target.delete()
                    return@withContext ApkDownloadResult.Failed(DownloadFailure.SIZE_MISMATCH)
                }
                if (!verifyDownloadedApk(target, sizeToVerify)) {
                    target.delete()
                    return@withContext ApkDownloadResult.Failed(DownloadFailure.CORRUPT)
                }
                ApkDownloadResult.Success(target)
            }
        } catch (e: Exception) {
            // Includes IOException from cancelDownload() and a failed
            // fsync: the partial file goes, never to be installed.
            target.delete()
            ApkDownloadResult.Failed(DownloadFailure.NETWORK)
        } finally {
            downloadCall = null
        }
    }

    /** Cancels an in-flight [downloadApk]; the partial file is deleted. */
    fun cancelDownload() {
        downloadCall?.cancel()
    }

    /** True when the system will let this app install an APK right now. */
    fun canInstallUnknownSources(): Boolean {
        return Build.VERSION.SDK_INT < Build.VERSION_CODES.O ||
            appContext.packageManager.canRequestPackageInstalls()
    }

    /**
     * Builds the system-installer intent for a downloaded APK, or null
     * when it cannot be built safely. The file is re-verified here
     * (exists, non-empty, expected size, ZIP magic) so a corrupt file is
     * never handed to the installer. ACTION_INSTALL_PACKAGE — not a
     * generic ACTION_VIEW — tells the system this is a package install,
     * which is what the installer expects; the MIME type is
     * application/vnd.android.package-archive and the FileProvider URI
     * carries FLAG_GRANT_READ_URI_PERMISSION so the installer can read
     * the file. FLAG_ACTIVITY_NEW_TASK is only added when the caller is
     * not an Activity: an Activity launching its own installer result
     * must not use it, or the result callback would never come back.
     *
     * The authority is "${packageName}.fileprovider" — exactly the
     * authority declared in AndroidManifest.xml — and the file must live
     * under filesDir/updates/, the <files-path> served in
     * res/xml/file_paths.xml. A mismatch there throws
     * IllegalArgumentException from getUriForFile and surfaces here as
     * null (install not started, file kept for retry) instead of a crash
     * or a "problem parsing the package" error from the installer.
     */
    fun buildInstallIntent(context: Context, apkFile: File, expectedSize: Long = -1): Intent? {
        if (!verifyDownloadedApk(apkFile, expectedSize)) return null
        return try {
            val uri = FileProvider.getUriForFile(
                context,
                "${context.packageName}.fileprovider",
                apkFile
            )
            Intent(Intent.ACTION_INSTALL_PACKAGE).apply {
                setDataAndType(uri, APK_MIME_TYPE)
                addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION)
                if (context !is android.app.Activity) {
                    addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
                }
            }
        } catch (e: Exception) {
            // IllegalArgumentException = file outside the FileProvider
            // paths (path mismatch); anything else = no installer.
            null
        }
    }

    /**
     * Fire-and-forget install, kept for callers without an
     * ActivityResultLauncher. Prefer [buildInstallIntent] + a launcher
     * so the outcome is known. Returns false when the installer could
     * not be started; the file is left in place for retry.
     */
    fun installApk(context: Context, apkFile: File, expectedSize: Long = -1): Boolean {
        val intent = buildInstallIntent(context, apkFile, expectedSize) ?: return false
        return try {
            context.startActivity(intent)
            true
        } catch (e: Exception) {
            false
        }
    }

    /**
     * Records that an install was started for [versionName]. The record
     * (with the download URL and asset size, so a retry can re-download
     * without another API check) lives in SharedPreferences and survives
     * process death — which matters because a successful install kills
     * and replaces this process, so the in-memory state is gone by the
     * time the new version runs. Cleared by [resolvePendingInstall]
     * once the outcome is known.
     */
    fun markInstallStarted(versionName: String, downloadUrl: String, assetSize: Long) {
        appContext.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE).edit()
            .putString(PREF_PENDING_VERSION, versionName)
            .putString(PREF_PENDING_URL, downloadUrl)
            .putLong(PREF_PENDING_SIZE, assetSize)
            .apply()
    }

    /** Clears the pending-install record (the APK itself is untouched). */
    fun clearPendingInstall() {
        appContext.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE).edit()
            .remove(PREF_PENDING_VERSION)
            .remove(PREF_PENDING_URL)
            .remove(PREF_PENDING_SIZE)
            .apply()
    }

    /** The pending install, if one was recorded and its APK is still on disk. */
    fun pendingInstall(): PendingInstall? {
        val prefs = appContext.getSharedPreferences(PREFS_NAME, Context.MODE_PRIVATE)
        val version = prefs.getString(PREF_PENDING_VERSION, null)
            ?.takeIf { it.isNotBlank() } ?: return null
        val file = downloadedApkFile()
        if (!file.isFile) {
            // APK gone (e.g. user cleared data): nothing to retry with.
            clearPendingInstall()
            return null
        }
        return PendingInstall(
            versionName = version,
            downloadUrl = prefs.getString(PREF_PENDING_URL, "") ?: "",
            assetSize = prefs.getLong(PREF_PENDING_SIZE, -1),
            file = file
        )
    }

    /**
     * Settles a pending install on Start-screen (re)entry: if the pending
     * version is now installed the APK is deleted and the record cleared;
     * if the old version is still installed the APK is kept so the user
     * can retry without downloading again. Call this on launch, on
     * resume, and after the installer returns.
     */
    fun resolvePendingInstall(): PendingInstallResult {
        val pending = pendingInstall() ?: return PendingInstallResult.None
        return if (isVersionInstalled(pending.versionName)) {
            deleteDownloadedApk()
            clearPendingInstall()
            PendingInstallResult.Installed
        } else {
            PendingInstallResult.StillPending(pending)
        }
    }

    /** True when the installed version is the target version or newer. */
    fun isVersionInstalled(versionName: String): Boolean {
        val installed = installedVersion().versionName
        val compared = compareSemver(installed, versionName)
        if (compared != null) return compared >= 0
        return installed == versionName
    }

    /**
     * Opens the system "install unknown apps" settings page for this app
     * (Android 8+). The caller keeps its dialog open, so when the user
     * comes back they can tap Install again and the permission is
     * re-checked at that moment.
     */
    fun openUnknownSourcesSettings(context: Context) {
        val intent = if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            Intent(Settings.ACTION_MANAGE_UNKNOWN_APP_SOURCES).apply {
                data = Uri.parse("package:${context.packageName}")
            }
        } else {
            Intent(Settings.ACTION_SECURITY_SETTINGS)
        }.apply {
            addFlags(Intent.FLAG_ACTIVITY_NEW_TASK)
        }
        try {
            context.startActivity(intent)
        } catch (e: Exception) {
            // No settings app to open: the dialog's Install button stays
            // as the retry path and will keep reporting the missing grant.
        }
    }

    companion object {
        const val RELEASES_URL =
            "https://api.github.com/repos/Kaia-Alenia/Axon/releases/latest"
        const val APK_ASSET_NAME = "axon-android-debug.apk"
        const val APK_FILE_NAME = "axon-update.apk"
        const val UPDATES_DIR_NAME = "updates"
        const val APK_MIME_TYPE = "application/vnd.android.package-archive"
        const val USER_AGENT = "Axon-Android-Updater"
        const val CHECK_TIMEOUT_SECONDS = 10L
        const val DOWNLOAD_BUFFER_BYTES = 8 * 1024
        const val PREFS_NAME = "axon_prefs"
        const val PREF_PENDING_VERSION = "pending_update_version"
        const val PREF_PENDING_URL = "pending_update_url"
        const val PREF_PENDING_SIZE = "pending_update_size"
    }
}

/** Installed app version, read from PackageManager. */
data class InstalledVersion(val versionName: String, val versionCode: Long)

/** A download whose install was started but whose outcome is unknown. */
data class PendingInstall(
    val versionName: String,
    val downloadUrl: String,
    val assetSize: Long,
    val file: File
)

/** Outcome of settling a pending install (see [AppUpdater.resolvePendingInstall]). */
sealed interface PendingInstallResult {
    /** Nothing was pending. */
    data object None : PendingInstallResult

    /** The pending version is now installed; the APK was deleted. */
    data object Installed : PendingInstallResult

    /** Still on the old version; the APK was kept for retry. */
    data class StillPending(val pending: PendingInstall) : PendingInstallResult
}

/** Why a download failed; drives the retry message shown in the UI. */
enum class DownloadFailure {
    /** Network error, bad HTTP status, cancelled, or I/O failure. */
    NETWORK,

    /** Bytes on disk do not equal the expected size (truncated download). */
    SIZE_MISMATCH,

    /** Neither the GitHub asset size nor Content-Length was known. */
    UNVERIFIABLE,

    /** Size matched but the file is not a plausible APK (bad ZIP magic). */
    CORRUPT,

    /** The file failed verification at install time, or the install
        intent could not be built (e.g. FileProvider path mismatch). */
    INSTALL_FAILED
}

/** Result of [AppUpdater.downloadApk]. Never throws. */
sealed interface ApkDownloadResult {
    /** The APK was downloaded, fsynced and fully verified. */
    data class Success(val file: File) : ApkDownloadResult

    /** The download failed; any partial file was deleted. */
    data class Failed(val reason: DownloadFailure) : ApkDownloadResult
}

/** Result of an update check. Never throws. */
sealed interface UpdateCheckResult {
    /** A newer release exists; [versionName] has no leading "v".
     * [assetSize] is the APK size in bytes from the GitHub API, or -1
     * when the API did not report one. */
    data class UpdateAvailable(
        val versionName: String,
        val downloadUrl: String,
        val assetSize: Long = -1
    ) : UpdateCheckResult

    /** Installed version is current (or the release tag matches it). */
    data object UpToDate : UpdateCheckResult

    /** Network error, bad response, rate limit, or no APK asset. */
    data object CheckFailed : UpdateCheckResult
}

/**
 * Parses "1.6.9" / "v1.7.0" / "1.7.0-rc1" into numeric components.
 * Returns null when a part is not numeric — a tag like "nightly" has no
 * version to compare against.
 */
fun parseSemver(version: String): List<Int>? {
    val withoutPrefix = version.trim().removePrefix("v").removePrefix("V")
    val core = withoutPrefix.split("-", "+").first()
    if (core.isEmpty()) return null
    val numbers = ArrayList<Int>()
    for (part in core.split(".")) {
        numbers.add(part.toIntOrNull() ?: return null)
    }
    return numbers.ifEmpty { null }
}

/**
 * Compares two version names as semver: negative when [a] is older,
 * 0 when equal, positive when newer, null when either side is unparsable.
 * Missing trailing components count as 0, so "1.7" equals "1.7.0".
 */
fun compareSemver(a: String, b: String): Int? {
    val partsA = parseSemver(a) ?: return null
    val partsB = parseSemver(b) ?: return null
    val length = maxOf(partsA.size, partsB.size)
    for (i in 0 until length) {
        val x = partsA.getOrElse(i) { 0 }
        val y = partsB.getOrElse(i) { 0 }
        if (x != y) return x.compareTo(y)
    }
    return 0
}

/**
 * True when the release [remoteTag] (e.g. "v1.7.0") is newer than the
 * installed [installedVersionName] (e.g. "1.6.9").
 *
 * Note on the versionCode fallback: the GitHub release payload carries no
 * remote versionCode, so an unparsable tag has nothing trustworthy to
 * compare against — it is treated as "no update" rather than guessing.
 * The installed versionCode is still read (see [InstalledVersion]) and
 * is available to callers for diagnostics.
 */
fun isNewerVersion(remoteTag: String, installedVersionName: String): Boolean {
    return (compareSemver(remoteTag, installedVersionName) ?: return false) > 0
}
