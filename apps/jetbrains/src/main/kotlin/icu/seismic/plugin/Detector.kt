package icu.seismic.plugin

import java.io.File
import com.intellij.openapi.application.ApplicationNamesInfo
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import java.net.InetAddress
import java.util.TimeZone
import java.util.concurrent.TimeUnit

object Detector {
    fun detectProject(project: Project): String = project.name

    /** Shells out to git directly — works regardless of whether the IDE's
     * own VCS integration is configured for the project. */
    fun detectBranch(project: Project): String? {
        val basePath = project.basePath ?: return null
        return try {
            val process = ProcessBuilder("git", "rev-parse", "--abbrev-ref", "HEAD")
                .directory(File(basePath))
                .redirectErrorStream(true)
                .start()
            if (!process.waitFor(2, TimeUnit.SECONDS)) {
                process.destroyForcibly()
                return null
            }
            if (process.exitValue() != 0) return null
            process.inputStream.bufferedReader().readText().trim().takeIf { it.isNotEmpty() }
        } catch (_: Exception) {
            null // git not installed, not a repo, etc — just skip branch info
        }
    }

    fun detectOS(): String {
        val name = System.getProperty("os.name").lowercase()
        return when {
            name.contains("win") -> "windows"
            name.contains("mac") -> "darwin"
            else -> "linux"
        }
    }

    fun detectMachine(): String = try {
        InetAddress.getLocalHost().hostName
    } catch (_: Exception) {
        "unknown"
    }

    fun detectTimezone(): String = TimeZone.getDefault().id

    /** Returns the IDE's own product name — e.g. "WebStorm", "GoLand",
     * "IntelliJ IDEA", "RustRover" — same idea as VS Code's editor: 'vscode',
     * but reflects whichever specific IDE is actually running. */
    fun detectEditorName(): String = ApplicationNamesInfo.getInstance().fullProductName

    fun shouldTrack(file: VirtualFile?): Boolean {
        if (file == null) return false
        if (!file.isInLocalFileSystem) return false
        if (!file.exists()) return false
        return true
    }

    fun languageId(file: VirtualFile): String = file.fileType.name.lowercase()
}