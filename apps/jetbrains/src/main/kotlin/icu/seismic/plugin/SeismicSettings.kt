package icu.seismic.plugin

import com.intellij.ide.util.PropertiesComponent

private const val KEY_API_KEY = "seismic.apiKey"
private const val KEY_API_URL = "seismic.apiUrl"
private const val KEY_ENABLED = "seismic.enabled"
private const val KEY_STATUS_BAR_ENABLED = "seismic.statusBarEnabled"

object SeismicSettings {
    private val props get() = PropertiesComponent.getInstance()

    fun getApiKey(): String {
        val local = props.getValue(KEY_API_KEY, "")
        if (local.isNotEmpty()) return local
        return GlobalConfigStore.read()?.apiKey ?: ""
    }

    fun getApiUrl(): String {
        val local = props.getValue(KEY_API_URL, "")
        if (local.isNotEmpty()) return local
        return GlobalConfigStore.read()?.apiUrl ?: ApiUrls.getDefaultApiUrl()
    }

    fun isEnabled(): Boolean = props.getBoolean(KEY_ENABLED, true)
    fun isStatusBarEnabled(): Boolean = props.getBoolean(KEY_STATUS_BAR_ENABLED, true)
    fun hasApiKey(): Boolean = getApiKey().trim().isNotEmpty()
    fun setEnabled(value: Boolean) = props.setValue(KEY_ENABLED, value, true)

    /**
     * Saves the API key in this IDE's local settings, and also writes it to
     * the shared ~/.seismic.cfg file so other editors pick it up automatically.
     */
    fun setApiKey(key: String) {
        props.setValue(KEY_API_KEY, key)
        GlobalConfigStore.write(GlobalConfig(apiKey = key, apiUrl = getApiUrl()))
    }

    fun setApiUrl(url: String) {
        props.setValue(KEY_API_URL, url)
    }

    fun setStatusBarEnabled(value: Boolean) {
        props.setValue(KEY_STATUS_BAR_ENABLED, value, true)
    }
}
