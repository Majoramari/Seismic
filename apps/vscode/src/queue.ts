import type { HeartbeatPayload } from './heartbeat';

/**
 * Stores heartbeats that failed to send (e.g. no internet).
 * They get retried later instead of being lost.
 */

interface QueuedHeartbeat {
  payload: HeartbeatPayload;
  attempts: number;
}

const MAX_SIZE = 100;
const MAX_ATTEMPTS = 3;

export class HeartbeatQueue {
  private queue: QueuedHeartbeat[] = [];

  enqueue(payload: HeartbeatPayload): void {
    if (this.queue.length >= MAX_SIZE) {
      this.queue.shift(); // drop the oldest one
    }
    this.queue.push({ payload, attempts: 0 });
  }

  size(): number {
    return this.queue.length;
  }

  /**
   * Tries to resend every queued heartbeat.
   * Successful ones are removed. Failed ones get another
   * attempt next time, up to MAX_ATTEMPTS before being dropped.
   */
  async flush(apiKey: string, apiUrl: string): Promise<void> {
    if (this.queue.length === 0 || !apiKey) return;

    const stillQueued: QueuedHeartbeat[] = [];

    for (const item of this.queue) {
      const success = await this.trySend(item.payload, apiKey, apiUrl);
      if (!success) {
        item.attempts++;
        if (item.attempts < MAX_ATTEMPTS) {
          stillQueued.push(item);
        }
      }
    }

    this.queue = stillQueued;
  }

  private async trySend(
    payload: HeartbeatPayload,
    apiKey: string,
    apiUrl: string,
  ): Promise<boolean> {
    try {
      const res = await fetch(`${apiUrl}/api/heartbeat`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify(payload),
      });
      return res.ok;
    } catch {
      return false;
    }
  }
}
