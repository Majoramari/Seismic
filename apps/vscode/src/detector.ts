import * as os from 'node:os';
import { execFile } from 'node:child_process';
import * as fs from 'node:fs/promises';
import * as path from 'node:path';
import { promisify } from 'node:util';
import type * as vscode from 'vscode';

/**
 * Pure functions that figure out information about the
 * current editor state: what file is open, what project,
 * what git branch, etc. Nothing here sends network requests.
 */

const execFileAsync = promisify(execFile);

function getVscode(): typeof vscode {
  return require('vscode') as typeof vscode;
}

export interface ProjectMetadata {
  project?: string;
  repoUrl?: string;
  websiteUrl?: string;
  lastCommitAt?: number;
  commits?: GitCommit[];
}

export interface GitCommit {
  hash: string;
  message?: string;
  authorName?: string;
  authorEmail?: string;
  committedAt?: number;
}

export function detectProject(document: vscode.TextDocument): string {
  const folder = getVscode().workspace.getWorkspaceFolder(document.uri);
  if (folder) return folder.name;

  // No workspace open — fall back to the file's own name
  const parts = document.fileName.split(/[\\/]/);
  return parts[parts.length - 1] || 'unknown';
}

export async function detectBranch(): Promise<string | undefined> {
  try {
    const gitExtension = getVscode().extensions.getExtension('vscode.git');
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

export async function detectProjectMetadata(document: vscode.TextDocument): Promise<ProjectMetadata> {
  const workspaceFolder = getVscode().workspace.getWorkspaceFolder(document.uri);
  if (!workspaceFolder) return {};

  return detectWorkspaceProjectMetadata(workspaceFolder);
}

export async function detectWorkspaceProjectMetadata(
  workspaceFolder: vscode.WorkspaceFolder,
): Promise<ProjectMetadata> {
  const workspacePath = workspaceFolder.uri.fsPath;
  const [repoUrl, websiteUrl, commits] = await Promise.all([
    detectRepoUrl(workspacePath),
    detectWebsiteUrl(workspacePath),
    detectRecentCommits(workspacePath),
  ]);
  const lastCommitAt = commits[0]?.committedAt;

  return { project: workspaceFolder.name, repoUrl, websiteUrl, lastCommitAt, commits };
}

async function detectRepoUrl(workspacePath: string): Promise<string | undefined> {
  try {
    const { stdout } = await execFileAsync('git', ['-C', workspacePath, 'config', '--get', 'remote.origin.url']);
    return normalizeGitUrl(stdout.trim());
  } catch {
    return undefined;
  }
}

async function detectRecentCommits(workspacePath: string, limit = 20): Promise<GitCommit[]> {
  try {
    const { stdout } = await execFileAsync('git', [
      '-C',
      workspacePath,
      'log',
      `-${limit}`,
      '--format=%H%x1f%cI%x1f%an%x1f%ae%x1f%s%x1e',
    ]);

    return stdout
      .split('\x1e')
      .map((record) => record.trim())
      .filter(Boolean)
      .map((record) => {
        const [hash, committedAtValue, authorName, authorEmail, message] = record.split('\x1f');
        const committedAt = Date.parse(committedAtValue);
        return {
          hash,
          committedAt: Number.isNaN(committedAt) ? undefined : committedAt,
          authorName: authorName || undefined,
          authorEmail: authorEmail || undefined,
          message: message || undefined,
        };
      })
      .filter((commit) => Boolean(commit.hash));
  } catch {
    return [];
  }
}

async function detectWebsiteUrl(workspacePath: string): Promise<string | undefined> {
  const packageJsonPath = await findNearestFile(workspacePath, 'package.json');
  if (!packageJsonPath) return undefined;

  try {
    const json = JSON.parse(await fs.readFile(packageJsonPath, 'utf8')) as {
      homepage?: unknown;
      website?: unknown;
    };
    if (typeof json.homepage === 'string' && json.homepage.trim()) return json.homepage.trim();
    if (typeof json.website === 'string' && json.website.trim()) return json.website.trim();
  } catch {
    return undefined;
  }

  return undefined;
}

async function findNearestFile(startPath: string, fileName: string): Promise<string | undefined> {
  let current = startPath;

  while (true) {
    const candidate = path.join(current, fileName);
    try {
      await fs.access(candidate);
      return candidate;
    } catch {
      const parent = path.dirname(current);
      if (parent === current) return undefined;
      current = parent;
    }
  }
}

function normalizeGitUrl(url: string): string | undefined {
  if (!url) return undefined;
  if (url.startsWith('git@')) {
    const match = url.match(/^git@([^:]+):(.+)$/);
    if (!match) return url;
    return `https://${match[1]}/${match[2].replace(/\.git$/, '')}`;
  }

  return url.replace(/\.git$/, '');
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
