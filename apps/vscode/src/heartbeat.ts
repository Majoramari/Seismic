import * as vscode from 'vscode';
import * as config from './config';
import * as detector from './detector';
import { HeartbeatQueue } from './queue';

export interface HeartbeatPayload {
    file: string;
    project: string;
    language: string;
    editor: 'vscode';
    branch?: string;
    repoUrl?: string;
    websiteUrl?: string;
    lastCommitAt?: number;
    os?: string;
    machine?: string;
    lines?: number;
    cursorLine?: number;
    timezone?: string;
    keystrokes: number;
    time: number;
}

const HEARTBEAT_INTERVAL_MS = 30 * 1000;

export class HeartbeatService {
    private queue = new HeartbeatQueue();
    private hasShownInvalidKeyWarning = false;

    // Characters inserted since the last heartbeat was sent.
    private keystrokeCount = 0;

    // The document to report on next tick, and whether anything has
    // happened since the last send. Every editor event just updates
    // these — it never triggers a network call directly. Only the
    // periodic tick() below decides when an actual heartbeat goes out,
    // so no matter how many times you switch tabs, save, or type, at
    // most one heartbeat fires per 30-second interval.
    private pendingDocument: vscode.TextDocument | null = null;
    private hasPendingActivity = false;

    private timer: ReturnType<typeof setInterval> | undefined;

    /** Called from onDidChangeTextDocument with characters inserted in that edit. */
    recordKeystrokes(charsInserted: number): void {
        if (charsInserted <= 0) return;
        this.keystrokeCount += charsInserted;
    }

    /**
     * Called from every relevant editor event (typing, switching files,
     * saving, gaining focus). Just records what's being worked on right
     * now — sending is entirely decided by the timer in start().
     */
    noteActivity(document: vscode.TextDocument): void {
        if (!config.isEnabled()) return;
        if (!detector.shouldTrack(document)) return;

        this.pendingDocument = document;
        this.hasPendingActivity = true;
    }

    /** Starts the periodic flush loop. Call once from activate(). */
    start(): void {
        this.timer = setInterval(() => {
            void this.tick();
        }, HEARTBEAT_INTERVAL_MS);
    }

    dispose(): void {
        if (this.timer) clearInterval(this.timer);
    }

    private async tick(): Promise<void> {
        if (!config.hasApiKey()) return;
        if (!this.hasPendingActivity || !this.pendingDocument) return;

        const document = this.pendingDocument;
        this.hasPendingActivity = false;

        const payload = await this.buildPayload(document);
        this.keystrokeCount = 0; // captured into payload above, start counting fresh
        await this.send(payload);
    }

    private async buildPayload(document: vscode.TextDocument): Promise<HeartbeatPayload> {
        const editor = vscode.window.activeTextEditor;

        const useGitRootProjectName = config.useGitRootProjectName();
        const metadata = await detector.detectProjectMetadata(document, useGitRootProjectName);

        return {
            file: document.fileName,
            project: await detector.detectProjectName(document, useGitRootProjectName),
            language: document.languageId,
            editor: 'vscode',
            branch: await detector.detectBranch(),
            repoUrl: metadata.repoUrl,
            websiteUrl: metadata.websiteUrl,
            lastCommitAt: metadata.lastCommitAt,
            os: detector.detectOS(),
            machine: detector.detectMachine(),
            lines: document.lineCount,
            cursorLine: editor ? editor.selection.active.line + 1 : undefined,
            timezone: detector.detectTimezone(),
            keystrokes: this.keystrokeCount,
            time: Date.now(),
        };
    }

    private async send(payload: HeartbeatPayload): Promise<void> {
        const apiKey = config.getApiKey();
        const apiUrl = config.getApiUrl();

        try {
            const res = await fetch(`${apiUrl}/api/heartbeat`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    Authorization: `Bearer ${apiKey}`,
                },
                body: JSON.stringify(payload),
            });

            if (res.status === 401) {
                this.notifyInvalidKey();
                return;
            }

            if (res.ok) {
                // Since we're online right now, also try to clear
                // out anything stuck in the offline queue.
                await this.queue.flush(apiKey, apiUrl);
            } else {
                this.queue.enqueue(payload);
            }
        } catch {
            // Network error (offline, DNS failure, etc) — queue it
            // silently and try again later. Never bother the user.
            this.queue.enqueue(payload);
        }
    }

    private notifyInvalidKey(): void {
        if (this.hasShownInvalidKeyWarning) return;
        this.hasShownInvalidKeyWarning = true;
        vscode.window.showWarningMessage(
            'Seismic: Invalid API key. Run "Seismic: Set API Key" to update it.',
        );
    }

    /**
     * Called periodically in the background to retry any
     * heartbeats that failed to send earlier.
     */
    async flushQueue(): Promise<void> {
        await this.queue.flush(config.getApiKey(), config.getApiUrl());
    }
}
