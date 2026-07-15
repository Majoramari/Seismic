package icu.seismic.plugin.startup

import com.intellij.openapi.application.ApplicationActivationListener
import com.intellij.openapi.application.ApplicationManager
import com.intellij.openapi.editor.event.EditorFactoryEvent
import com.intellij.openapi.editor.event.EditorFactoryListener
import com.intellij.openapi.editor.EditorFactory
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
import java.awt.event.KeyAdapter
import java.awt.event.KeyEvent
import java.util.concurrent.Executors
import java.util.concurrent.TimeUnit

class SeismicStartupActivity : ProjectActivity, DumbAware {
    override suspend fun execute(project: Project) {
        val heartbeat = HeartbeatService.getInstance()
        heartbeat.syncProjectMetadata(project, forced = true)

        // Heartbeat on typing (subject to the 30-second throttle). Uses a raw
        // AWT key listener rather than a DocumentListener, since
        // DocumentEvent fires for ANY text change — auto-format,
        // code completion, auto-import, refactors — not just real
        // keypresses. A key listener only fires for actual keys the
        // user physically pressed.
        EditorFactory.getInstance().addEditorFactoryListener(object : EditorFactoryListener {
            override fun editorCreated(event: EditorFactoryEvent) {
                val editor = event.editor
                editor.contentComponent.addKeyListener(object : KeyAdapter() {
                    override fun keyTyped(e: KeyEvent) {
                        heartbeat.recordKeystrokes(1)

                        val file = FileDocumentManager.getInstance().getFile(editor.document) ?: return
                        val currentProject = editor.project ?: return
                        val selectedEditor = FileEditorManager.getInstance(currentProject).selectedTextEditor
                        heartbeat.handleActivity(currentProject, file, selectedEditor)
                    }
                })
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
