package icu.seismic.plugin.startup

import com.intellij.openapi.application.ApplicationActivationListener
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.editor.EditorFactory
import com.intellij.openapi.editor.event.DocumentEvent
import com.intellij.openapi.editor.event.DocumentListener
import com.intellij.openapi.fileEditor.FileDocumentManager
import com.intellij.openapi.fileEditor.FileDocumentManagerListener
import com.intellij.openapi.fileEditor.FileEditorManager
import com.intellij.openapi.fileEditor.FileEditorManagerEvent
import com.intellij.openapi.fileEditor.FileEditorManagerListener
import com.intellij.openapi.project.DumbAware
import com.intellij.openapi.project.Project
import com.intellij.openapi.startup.ProjectActivity
import com.intellij.openapi.wm.IdeFrame
import icu.seismic.plugin.HeartbeatService
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit

class SeismicStartupActivity : ProjectActivity, DumbAware {
    override suspend fun execute(project: Project) {
        val heartbeat = HeartbeatService.getInstance()

        // Heartbeat on typing (subject to the 2 min rule)
        EditorFactory.getInstance().eventMulticaster.addDocumentListener(object : DocumentListener {
            override fun documentChanged(event: DocumentEvent) {
                heartbeat.recordKeystrokes(event.newFragment.length)
                val file = FileDocumentManager.getInstance().getFile(event.document) ?: return
                val editor = FileEditorManager.getInstance(project).selectedTextEditor
                heartbeat.handleActivity(project, file, editor)
            }
        }, project)

        val connection = project.messageBus.connect()

        // Heartbeat immediately on switching files
        connection.subscribe(FileEditorManagerListener.FILE_EDITOR_MANAGER, object : FileEditorManagerListener {
            override fun selectionChanged(event: FileEditorManagerEvent) {
                val file = event.newFile ?: return
                val editor = FileEditorManager.getInstance(project).selectedTextEditor
                heartbeat.handleActivity(project, file, editor, forced = true)
            }
        })

        // Heartbeat immediately on save
        connection.subscribe(FileDocumentManagerListener.TOPIC, object : FileDocumentManagerListener {
            override fun beforeDocumentSaving(document: com.intellij.openapi.editor.Document) {
                val file = FileDocumentManager.getInstance().getFile(document) ?: return
                val editor = FileEditorManager.getInstance(project).selectedTextEditor
                heartbeat.handleActivity(project, file, editor, forced = true)
            }
        })

        // Heartbeat when the IDE window gains focus
        ApplicationManager.getApplication().messageBus.connect(project)
            .subscribe(ApplicationActivationListener.TOPIC, object : ApplicationActivationListener {
                override fun applicationActivated(ideFrame: IdeFrame) {
                    val editor = FileEditorManager.getInstance(project).selectedTextEditor ?: return
                    val file = FileDocumentManager.getInstance().getFile(editor.document) ?: return
                    heartbeat.handleActivity(project, file, editor) // respect throttle
                }
            })

        // Retry queued heartbeats every 5 minutes
        val scheduler = Executors.newSingleThreadScheduledExecutor { r ->
            Thread(r, "Seismic-FlushQueue").apply { isDaemon = true }
        }
        scheduler.scheduleWithFixedDelay({ heartbeat.flushQueue() }, 5, 5, TimeUnit.MINUTES)
    }
}