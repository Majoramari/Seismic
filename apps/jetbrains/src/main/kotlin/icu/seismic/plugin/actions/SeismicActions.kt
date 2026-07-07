package icu.seismic.plugin.actions

import com.intellij.ide.BrowserUtil
import com.intellij.openapi.actionSystem.AnAction
import com.intellij.openapi.actionSystem.AnActionEvent
import com.intellij.openapi.project.Project
import com.intellij.openapi.ui.Messages
import icu.seismic.plugin.SeismicSettings

fun promptForApiKey(project: Project?) {
    val key = Messages.showPasswordDialog(
        project,
        "Find your API key at seismic.icu/settings",
        "Enter your Seismic API key",
        null
    )
    if (!key.isNullOrBlank()) {
        SeismicSettings.setApiKey(key)
        Messages.showInfoMessage(project, "API key saved!", "Seismic")
    }
}

class SetApiKeyAction : AnAction("Set API Key...") {
    override fun actionPerformed(e: AnActionEvent) = promptForApiKey(e.project)
}

class OpenDashboardAction : AnAction("Open Dashboard") {
    override fun actionPerformed(e: AnActionEvent) = BrowserUtil.browse("https://seismic.icu/dashboard")
}

class EnableTrackingAction : AnAction("Enable Tracking") {
    override fun actionPerformed(e: AnActionEvent) {
        SeismicSettings.setEnabled(true)
        Messages.showInfoMessage(e.project, "Tracking enabled", "Seismic")
    }
}

class DisableTrackingAction : AnAction("Disable Tracking") {
    override fun actionPerformed(e: AnActionEvent) {
        SeismicSettings.setEnabled(false)
        Messages.showInfoMessage(e.project, "Tracking disabled", "Seismic")
    }
}