import * as vscode from 'vscode';
import * as config from './config';
import * as detector from './detector';

interface ProjectSyncPayload {
  project: string;
  repoUrl?: string;
  websiteUrl?: string;
  lastCommitAt?: number;
  commits: detector.GitCommit[];
}

const MIN_SYNC_INTERVAL_MS = 15 * 1000;

export class ProjectSyncService {
  private lastSyncByFolder = new Map<string, number>();
  private hasShownInvalidKeyWarning = false;

  async syncWorkspaceFolders(force = false): Promise<void> {
    const folders = vscode.workspace.workspaceFolders ?? [];
    await Promise.all(folders.map((folder) => this.syncWorkspaceFolder(folder, force)));
  }

  async syncDocument(document: vscode.TextDocument, force = false): Promise<void> {
    if (!config.isEnabled()) return;
    if (!detector.shouldTrack(document)) return;

    const folder = vscode.workspace.getWorkspaceFolder(document.uri);
    if (!folder) return;

    await this.syncWorkspaceFolder(folder, force);
  }

  async syncWorkspaceFolder(folder: vscode.WorkspaceFolder, force = false): Promise<void> {
    if (!config.isEnabled()) return;
    if (!config.hasApiKey()) return;

    const folderPath = folder.uri.fsPath;
    const lastSync = this.lastSyncByFolder.get(folderPath) ?? 0;
    if (!force && Date.now() - lastSync < MIN_SYNC_INTERVAL_MS) return;

    const metadata = await detector.detectWorkspaceProjectMetadata(folder);
    const payload: ProjectSyncPayload = {
      project: metadata.project ?? folder.name,
      repoUrl: metadata.repoUrl,
      websiteUrl: metadata.websiteUrl,
      lastCommitAt: metadata.lastCommitAt,
      commits: metadata.commits ?? [],
    };

    if (!payload.repoUrl && payload.commits.length === 0 && !payload.websiteUrl) return;

    const ok = await this.send(payload);
    if (ok) this.lastSyncByFolder.set(folderPath, Date.now());
  }

  private async send(payload: ProjectSyncPayload): Promise<boolean> {
    const apiKey = config.getApiKey();
    const apiUrl = config.getApiUrl();

    try {
      const res = await fetch(`${apiUrl}/api/projects/sync`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${apiKey}`,
        },
        body: JSON.stringify(payload),
      });

      if (res.status === 401) {
        this.notifyInvalidKey();
        return false;
      }

      return res.ok;
    } catch {
      return false;
    }
  }

  private notifyInvalidKey(): void {
    if (this.hasShownInvalidKeyWarning) return;
    this.hasShownInvalidKeyWarning = true;
    vscode.window.showWarningMessage(
      'Seismic: Invalid API key. Run "Seismic: Set API Key" to update it.',
    );
  }
}
