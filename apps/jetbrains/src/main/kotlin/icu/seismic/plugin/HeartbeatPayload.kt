package icu.seismic.plugin

data class HeartbeatPayload(
    val file: String,
    val project: String,
    val language: String,
    val editor: String,
    val branch: String? = null,
    val os: String? = null,
    val machine: String? = null,
    val lines: Int? = null,
    val cursorLine: Int? = null,
    val timezone: String? = null,
    val keystrokes: Int = 0,
    val time: Long
)
