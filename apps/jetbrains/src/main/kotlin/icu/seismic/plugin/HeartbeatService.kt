package icu.seismic.plugin

import com.google.gson.Gson
import com.intellij.notification.NotificationGroupManager
import com.intellij.notification.NotificationType
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.editor.Editor
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody
import java.util.concurrent.TimeUnit
import java.util.concurrent.atomic.AtomicInteger

private const val TWO_MINUTES_MS = 2 * 60 * 1000L
private val JSON = "application/json".toMediaType()

private const val MAX_KEYSTROKE_EDIT_LENGTH = 200

class HeartbeatService {
    private val client = OkHttpClient.Builder()
        .connectTimeout(5, TimeUnit.SECONDS)
        .readTimeout(5, TimeUnit.SECONDS)
        .build()
    private val gson = Gson()
    private val queue = HeartbeatQueue(client, gson)

    @Volatile
    private var lastHeartbeatTime = 0L

    @Volatile
    private var lastFile: String? = null

    @Volatile
    private var hasShownInvalidKeyWarning = false

    // Characters inserted since the last heartbeat was sent. Must be
    // an instance member (not top-level) since HeartbeatService is
    // used as a per-application singleton via getInstance(), and
    // SeismicStartupActivity calls heartbeat.recordKeystrokes(...)
    // on that instance.
    private val keystrokeCount = AtomicInteger(0)

    /** Called from the document listener with characters inserted in that edit. */
    fun recordKeystrokes(charsInserted: Int) {
        if (charsInserted <= 0 || charsInserted > MAX_KEYSTROKE_EDIT_LENGTH) return
        keystrokeCount.addAndGet(charsInserted)
    }

    fun handleActivity(project: Project, file: VirtualFile, editor: Editor?, forced: Boolean = false) {
        if (!SeismicSettings.isEnabled()) return
        if (!SeismicSettings.hasApiKey()) return
        if (!Detector.shouldTrack(file)) return

        val now = System.currentTimeMillis()
        val fileChanged = file.path != lastFile

        // IMPORTANT: check the throttle BEFORE touching keystrokeCount.
        // If we're not actually going to send a heartbeat, we must not
        // reset the counter — otherwise typing gets wiped out between
        // real sends and the total is always near zero.
        if (!forced && !fileChanged && now - lastHeartbeatTime < TWO_MINUTES_MS) return

        lastHeartbeatTime = now
        lastFile = file.path

        // Capture what we need on the calling thread; do the slow work
        // (git subprocess, network I/O) on a background thread.
        val filePath = file.path
        val lineCount = editor?.document?.lineCount
        val cursorLine = editor?.caretModel?.logicalPosition?.line?.plus(1)
        // Read-and-reset now that we know a heartbeat will actually send.
        val keystrokes = keystrokeCount.getAndSet(0)

        ApplicationManager.getApplication().executeOnPooledThread {
            val payload = HeartbeatPayload(
                file = filePath,
                project = Detector.detectProject(project),
                language = Detector.languageId(file),
                editor = Detector.detectEditorName(),
                branch = Detector.detectBranch(project),
                os = Detector.detectOS(),
                machine = Detector.detectMachine(),
                lines = lineCount,
                cursorLine = cursorLine,
                timezone = Detector.detectTimezone(),
                keystrokes = keystrokes,
                time = System.currentTimeMillis()
            )
            send(payload)
        }
    }

    private fun send(payload: HeartbeatPayload) {
        val apiKey = SeismicSettings.getApiKey()
        val apiUrl = SeismicSettings.getApiUrl()

        try {
            val body = gson.toJson(payload).toRequestBody(JSON)
            val request = Request.Builder()
                .url("$apiUrl/api/heartbeat")
                .post(body)
                .addHeader("Authorization", "Bearer $apiKey")
                .build()

            client.newCall(request).execute().use { response ->
                when {
                    response.code == 401 -> notifyInvalidKey()
                    response.isSuccessful -> queue.flush(apiKey, apiUrl)
                    else -> queue.enqueue(payload)
                }
            }
        } catch (_: Exception) {
            queue.enqueue(payload)
        }
    }

    private fun notifyInvalidKey() {
        if (hasShownInvalidKeyWarning) return
        hasShownInvalidKeyWarning = true
        NotificationGroupManager.getInstance()
            .getNotificationGroup("Seismic Notifications")
            .createNotification(
                "Seismic: Invalid API key. Use Tools > Seismic > Set API Key to update it.",
                NotificationType.WARNING
            )
            .notify(null)
    }

    fun flushQueue() {
        queue.flush(SeismicSettings.getApiKey(), SeismicSettings.getApiUrl())
    }

    companion object {
        fun getInstance(): HeartbeatService =
            ApplicationManager.getApplication().getService(HeartbeatService::class.java)
    }
}
