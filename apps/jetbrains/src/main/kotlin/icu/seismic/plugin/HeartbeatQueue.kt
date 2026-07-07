package icu.seismic.plugin

import com.google.gson.Gson
import okhttp3.MediaType.Companion.toMediaType
import okhttp3.OkHttpClient
import okhttp3.Request
import okhttp3.RequestBody.Companion.toRequestBody

private const val MAX_SIZE = 100
private const val MAX_ATTEMPTS = 3
private val JSON = "application/json".toMediaType()

private data class QueuedHeartbeat(val payload: HeartbeatPayload, var attempts: Int = 0)

class HeartbeatQueue(private val client: OkHttpClient, private val gson: Gson) {
    private val queue = mutableListOf<QueuedHeartbeat>()

    @Synchronized
    fun enqueue(payload: HeartbeatPayload) {
        if (queue.size >= MAX_SIZE) queue.removeAt(0) // drop the oldest
        queue.add(QueuedHeartbeat(payload))
    }

    @Synchronized
    fun flush(apiKey: String, apiUrl: String) {
        if (queue.isEmpty() || apiKey.isEmpty()) return

        val stillQueued = mutableListOf<QueuedHeartbeat>()
        for (item in queue.toList()) {
            if (!trySend(item.payload, apiKey, apiUrl)) {
                item.attempts++
                if (item.attempts < MAX_ATTEMPTS) stillQueued.add(item)
            }
        }
        queue.clear()
        queue.addAll(stillQueued)
    }

    private fun trySend(payload: HeartbeatPayload, apiKey: String, apiUrl: String): Boolean {
        return try {
            val body = gson.toJson(payload).toRequestBody(JSON)
            val request = Request.Builder()
                .url("$apiUrl/api/heartbeat")
                .post(body)
                .addHeader("Authorization", "Bearer $apiKey")
                .build()
            client.newCall(request).execute().use { it.isSuccessful }
        } catch (_: Exception) {
            false
        }
    }
}