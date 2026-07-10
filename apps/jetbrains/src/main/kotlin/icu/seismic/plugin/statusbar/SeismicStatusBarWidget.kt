package icu.seismic.plugin.statusbar

import com.google.gson.Gson
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.project.Project
import com.intellij.openapi.util.IconLoader
import com.intellij.openapi.wm.CustomStatusBarWidget
import com.intellij.openapi.wm.StatusBar
import com.intellij.openapi.wm.StatusBarWidget
import com.intellij.util.Alarm
import com.intellij.util.ui.JBUI
import icu.seismic.plugin.SeismicSettings
import icu.seismic.plugin.actions.promptForApiKey
import okhttp3.OkHttpClient
import okhttp3.Request
import java.awt.Cursor
import java.awt.event.MouseAdapter
import java.awt.event.MouseEvent
import java.util.concurrent.TimeUnit
import javax.swing.JComponent
import javax.swing.JLabel
import javax.swing.SwingUtilities

private const val UPDATE_INTERVAL_MS = 60_000L

class SeismicStatusBarWidget(private val project: Project) : CustomStatusBarWidget {

    private val alarm = Alarm(Alarm.ThreadToUse.POOLED_THREAD, this)
    private val client = OkHttpClient.Builder()
        .connectTimeout(5, TimeUnit.SECONDS).readTimeout(5, TimeUnit.SECONDS).build()
    private val gson = Gson()

    private val logoIcon = IconLoader.getIcon("/icons/seismicStatusBar.svg", SeismicStatusBarWidget::class.java)

    private val label = JLabel("Seismic", logoIcon, JLabel.LEFT).apply {
        border = JBUI.Borders.empty(0, 6)
        cursor = Cursor.getPredefinedCursor(Cursor.HAND_CURSOR)
        iconTextGap = 4
        // Icon never changes — it's always the logo, regardless of state.
        // Only the text reflects whether a key is set / tracking is on / paused.
        addMouseListener(object : MouseAdapter() {
            override fun mouseClicked(e: MouseEvent) {
                if (!SeismicSettings.hasApiKey()) {
                    promptForApiKey(project)
                    refresh()
                    return
                }
                com.intellij.ide.BrowserUtil.browse("https://seismic.icu/dashboard")
            }
        })
    }

    override fun ID(): String = "SeismicStatusBarWidget"
    override fun getComponent(): JComponent = label
    override fun install(statusBar: StatusBar) {
        scheduleUpdate()
    }

    override fun dispose() {
        alarm.dispose()
    }

    override fun getPresentation(): StatusBarWidget.WidgetPresentation? = null

    private fun scheduleUpdate() {
        refresh()
        alarm.addRequest({ refresh(); scheduleUpdate() }, UPDATE_INTERVAL_MS)
    }

    fun refresh() {
        ApplicationManager.getApplication().executeOnPooledThread {
            val (newText, newTooltip) = computeContent()
            SwingUtilities.invokeLater {
                label.text = newText
                label.toolTipText = newTooltip
                label.icon = logoIcon // re-assert every refresh, never let it change
            }
        }
    }

    private fun computeContent(): Pair<String, String> {
        if (!SeismicSettings.isStatusBarEnabled()) return "" to ""

        if (!SeismicSettings.hasApiKey()) {
            return "Set API Key" to "Click to set your Seismic API key"
        }
        if (!SeismicSettings.isEnabled()) {
            return "Paused" to "Seismic tracking is disabled"
        }

        return try {
            val seconds = fetchTodaySeconds()
            formatSeconds(seconds) to "Today's coding time on Seismic\nClick to open dashboard"
        } catch (_: Exception) {
            "Offline" to "Could not connect to Seismic — click to open dashboard"
        }
    }

    private fun fetchTodaySeconds(): Long {
        val request = Request.Builder()
            .url("${SeismicSettings.getApiUrl()}/api/stats/summary?range=today")
            .addHeader("Authorization", "Bearer ${SeismicSettings.getApiKey()}")
            .build()
        client.newCall(request).execute().use { response ->
            if (!response.isSuccessful) throw Exception("failed to fetch stats")
            val json = gson.fromJson(response.body?.string(), StatsResponse::class.java)
            return json.data.totalSeconds
        }
    }

    private fun formatSeconds(seconds: Long): String {
        if (seconds < 60) return "< 1m"
        val hours = seconds / 3600
        val minutes = (seconds % 3600) / 60
        return if (hours > 0) "${hours}h ${minutes}m" else "${minutes}m"
    }

    private data class StatsResponse(val success: Boolean, val data: StatsData)
    private data class StatsData(val totalSeconds: Long)
}
