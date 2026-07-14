package icu.seismic.plugin

data class ProjectSyncPayload(
    val project: String,
    val repoUrl: String? = null,
    val websiteUrl: String? = null,
    val lastCommitAt: Long? = null,
    val commits: List<ProjectCommitPayload> = emptyList(),
)

data class ProjectCommitPayload(
    val hash: String,
    val message: String? = null,
    val authorName: String? = null,
    val authorEmail: String? = null,
    val committedAt: Long? = null,
)
