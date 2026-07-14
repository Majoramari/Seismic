package icu.seismic.plugin

import java.io.File
import com.intellij.openapi.application.ApplicationNamesInfo
import com.intellij.openapi.project.Project
import com.intellij.openapi.vfs.VirtualFile
import java.net.InetAddress
import java.util.TimeZone
import java.util.concurrent.TimeUnit

object Detector {
    fun detectProject(project: Project, useGitRoot: Boolean = false): String {
        if (useGitRoot) {
            detectGitRoot(project)?.let { return File(it).name }
        }
        return project.name
    }

    fun detectGitRoot(project: Project): String? {
        val basePath = project.basePath ?: return null
        return runGit(basePath, listOf("rev-parse", "--show-toplevel"))
    }

    /** Shells out to git directly — works regardless of whether the IDE's
     * own VCS integration is configured for the project. */
    fun detectBranch(project: Project): String? {
        val basePath = project.basePath ?: return null
        return try {
            runGit(basePath, listOf("rev-parse", "--abbrev-ref", "HEAD"))
        } catch (_: Exception) {
            null // git not installed, not a repo, etc — just skip branch info
        }
    }

    fun detectProjectMetadata(project: Project, useGitRoot: Boolean): ProjectSyncPayload? {
        val basePath = project.basePath ?: return null
        val projectPath = if (useGitRoot) detectGitRoot(project) ?: basePath else basePath
        val commits = detectCommits(projectPath)

        return ProjectSyncPayload(
            project = File(projectPath).name.ifEmpty { project.name },
            repoUrl = detectRepoUrl(projectPath),
            websiteUrl = detectWebsiteUrl(projectPath),
            lastCommitAt = commits.firstOrNull()?.committedAt,
            commits = commits,
        )
    }

    private fun detectRepoUrl(projectPath: String): String? {
        val url = runGit(projectPath, listOf("config", "--get", "remote.origin.url")) ?: return null
        val sshMatch = Regex("^git@([^:]+):(.+)$").find(url)
        if (sshMatch != null) {
            val host = sshMatch.groupValues[1]
            val repo = sshMatch.groupValues[2].removeSuffix(".git")
            return "https://$host/$repo"
        }
        return url.removeSuffix(".git")
    }

    private fun detectWebsiteUrl(projectPath: String): String? {
        val packageJson = File(projectPath, "package.json")
        if (!packageJson.isFile) return null

        val text = runCatching { packageJson.readText() }.getOrNull() ?: return null
        val homepage = Regex("\"homepage\"\\s*:\\s*\"([^\"]+)\"").find(text)?.groupValues?.get(1)
        val website = Regex("\"website\"\\s*:\\s*\"([^\"]+)\"").find(text)?.groupValues?.get(1)
        return homepage ?: website
    }

    private fun detectCommits(projectPath: String, limit: Int = 20): List<ProjectCommitPayload> {
        val output = runGit(projectPath, listOf("log", "-$limit", "--format=%H%x1f%ct%x1f%an%x1f%ae%x1f%s"))
            ?: return emptyList()

        return output.lineSequence()
            .filter { it.isNotBlank() }
            .mapNotNull { line ->
                val parts = line.split('\u001f')
                if (parts.isEmpty() || parts[0].isBlank()) return@mapNotNull null
                ProjectCommitPayload(
                    hash = parts[0],
                    committedAt = parts.getOrNull(1)?.toLongOrNull()?.times(1000),
                    authorName = parts.getOrNull(2),
                    authorEmail = parts.getOrNull(3),
                    message = parts.getOrNull(4),
                )
            }
            .toList()
    }

    private fun runGit(projectPath: String, args: List<String>): String? {
        return try {
            val process = ProcessBuilder(listOf("git", "-C", projectPath) + args)
                .redirectErrorStream(true)
                .start()
            if (!process.waitFor(2, TimeUnit.SECONDS)) {
                process.destroyForcibly()
                return null
            }
            if (process.exitValue() != 0) return null
            process.inputStream.bufferedReader().readText().trim().takeIf { it.isNotEmpty() }
        } catch (_: Exception) {
            null
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
