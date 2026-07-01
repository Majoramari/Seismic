import * as os from 'node:os';
import * as vscode from 'vscode';

/**
 * Pure functions that figure out information about the
 * current editor state: what file is open, what project,
 * what git branch, etc. Nothing here sends network requests.
 */

export function detectProject(document: vscode.TextDocument): string {
  const folder = vscode.workspace.getWorkspaceFolder(document.uri);
  if (folder) return folder.name;

  // No workspace open — fall back to the file's own name
  const parts = document.fileName.split(/[\\/]/);
  return parts[parts.length - 1] || 'unknown';
}

export async function detectBranch(): Promise<string | undefined> {
  try {
    const gitExtension = vscode.extensions.getExtension('vscode.git');
    if (!gitExtension) return undefined;

    const gitApi = gitExtension.isActive
      ? gitExtension.exports.getAPI(1)
      : (await gitExtension.activate()).getAPI(1);

    const repo = gitApi.repositories[0];
    return repo?.state?.HEAD?.name;
  } catch {
    // If anything goes wrong here, just skip the branch info
    // rather than crashing the extension.
    return undefined;
  }
}

export function detectOS(): string {
  return os.platform(); // "linux" | "win32" | "darwin"
}

export function detectMachine(): string {
  return os.hostname();
}

export function detectTimezone(): string {
  return Intl.DateTimeFormat().resolvedOptions().timeZone;
}

/**
 * Decides whether a given document is worth tracking at all.
 * Filters out untitled files and non-file schemes like
 * git:// or output:// that VS Code sometimes opens internally.
 */
export function shouldTrack(document: vscode.TextDocument): boolean {
  if (document.isUntitled) return false;
  if (document.uri.scheme !== 'file') return false;
  return true;
}
