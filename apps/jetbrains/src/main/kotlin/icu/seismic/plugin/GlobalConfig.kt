package icu.seismic.plugin

import java.io.File

data class GlobalConfig(val apiKey: String, val apiUrl: String)

object GlobalConfigStore {
    private val configFile = File(System.getProperty("user.home"), ".seismic.cfg")

    fun read(): GlobalConfig? {
        if (!configFile.exists()) return null

        val content = configFile.readText()
        val apiKey = Regex("api_key\\s*=\\s*(.+)").find(content)?.groupValues?.get(1)?.trim() ?: ""
        val apiUrl = Regex("api_url\\s*=\\s*(.+)").find(content)?.groupValues?.get(1)?.trim() ?: ""

        if (apiKey.isEmpty()) return null
        return GlobalConfig(apiKey, apiUrl.ifEmpty { ApiUrls.PUBLISHED_API_URL })
    }

    fun write(config: GlobalConfig) {
        val content = "[settings]\napi_key = ${config.apiKey}\napi_url = ${config.apiUrl}\n"
        configFile.writeText(content)
    }
}