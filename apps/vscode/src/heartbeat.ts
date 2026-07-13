import * as vscode from 'vscode';
import * as config from './config';
import * as detector from './detector';
import {HeartbeatQueue} from './queue';

export interface HeartbeatPayload {
    file: string;
    project: string;
    language: string;
    editor: 'vscode';
    branch?: string;
    os?: string;
    machine?: string;
    lines?: number;
    cursorLine?: number;
    timezone?: string;
    keystrokes: number;
    time: number;
}

const TWO_MINUTES = 2 * 60 * 1000;

export class HeartbeatService {
    private lastHeartbeatTime = 0;
    private lastFile = '';
    private queue = new HeartbeatQueue();
    private hasShownInvalidKeyWarning = false;
    private keystrokeCount = 0;

    recordKeystrokes(charsInserted: number): void {
        if (charsInserted <= 0) return;
        this.keystrokeCount += charsInserted;
    }

    async handleActivity(document: vscode.TextDocument, forced = false): Promise<void> {
        if (!config.isEnabled()) return;
        if (!config.hasApiKey()) return;
        if (!detector.shouldTrack(document)) return;

        const now = Date.now();
        const fileChanged = document.fileName !== this.lastFile;

        if (!forced && !fileChanged && now - this.lastHeartbeatTime < TWO_MINUTES) {
            return;
        }

        this.lastHeartbeatTime = now;
        this.lastFile = document.fileName;

        const payload = await this.buildPayload(document);
        this.keystrokeCount = 0;
        await this.send(payload);
    }

    private async buildPayload(document: vscode.TextDocument): Promise<HeartbeatPayload> {
        const editor = vscode.window.activeTextEditor;

        return {
            file: document.fileName,
            project: detector.detectProject(document),
            language: document.languageId,
            editor: 'vscode',
            branch: await detector.detectBranch(),
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
                await this.queue.flush(apiKey, apiUrl);
            } else {
                this.queue.enqueue(payload);
            }
        } catch {
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

    async flushQueue(): Promise<void> {
        await this.queue.flush(config.getApiKey(), config.getApiUrl());
    }
}
