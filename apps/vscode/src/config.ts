import * as vscode from 'vscode';
import { readGlobalConfig, writeGlobalConfig } from './globalConfig';

/**
 * Handles reading and writing Seismic settings from VS Code's
 * configuration system. All settings live under the "seismic"
 * namespace in the user's settings.json.
 *
 * The API key specifically also falls back to a shared config
 * file at ~/.seismic.cfg, so it can be set once and reused
 * automatically by other editors (Neovim, JetBrains) without
 * configuring each one separately.
 */

const SECTION = 'seismic';

export function getApiKey(): string {
  const vsCodeKey = vscode.workspace.getConfiguration(SECTION).get<string>('apiKey', '');
  if (vsCodeKey) return vsCodeKey;

  // No key set in VS Code settings — check the shared config
  // file that other editors might have already written to.
  const global = readGlobalConfig();
  return global?.apiKey ?? '';
}

export function getApiUrl(): string {
  const vsCodeUrl = vscode.workspace.getConfiguration(SECTION).get<string>('apiUrl', '');
  if (vsCodeUrl) return vsCodeUrl;

  const global = readGlobalConfig();
  return global?.apiUrl ?? 'https://api.seismic.icu';
}

export function isEnabled(): boolean {
  return vscode.workspace.getConfiguration(SECTION).get<boolean>('enabled', true);
}

export function isStatusBarEnabled(): boolean {
  return vscode.workspace.getConfiguration(SECTION).get<boolean>('statusBarEnabled', true);
}

export function hasApiKey(): boolean {
  return getApiKey().trim().length > 0;
}

/**
 * Saves the API key to VS Code's global settings (not per-workspace)
 * so it persists across every project the developer opens, and also
 * writes it to the shared ~/.seismic.cfg file so other editors
 * (Neovim, JetBrains) can pick it up automatically.
 */
export async function setApiKey(key: string): Promise<void> {
  await vscode.workspace
    .getConfiguration(SECTION)
    .update('apiKey', key, vscode.ConfigurationTarget.Global);

  writeGlobalConfig({ apiKey: key, apiUrl: getApiUrl() });
}
