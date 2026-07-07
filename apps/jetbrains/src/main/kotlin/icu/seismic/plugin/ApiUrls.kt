package icu.seismic.plugin

object ApiUrls {
    const val DEV_API_URL = "http://localhost:5024"
    const val PUBLISHED_API_URL = "https://correct-wolverine-majoramari-6049fd71.koyeb.app"

    fun getDefaultApiUrl(isDevelopment: Boolean = false): String =
        if (isDevelopment) DEV_API_URL else PUBLISHED_API_URL
}