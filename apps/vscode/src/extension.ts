import * as vscode from 'vscode';
import * as config from './config';
import { HeartbeatService } from './heartbeat';
import { StatusBarManager } from './statusbar';

/**
 * Entry point for the Seismic extension. VS Code calls
 * activate() once when the extension starts, and deactivate()
 * when it shuts down.
 */
export function activate(context: vscode.ExtensionContext) {
  const heartbeat = new HeartbeatService();
  const statusBar = new StatusBarManager();

  // Send a heartbeat when the user types (subject to the 2 min rule)
  context.subscriptions.push(
    vscode.workspace.onDidChangeTextDocument((e) => {
      heartbeat.handleActivity(e.document);
    }),
  );

  // Send a heartbeat immediately when switching files
  context.subscriptions.push(
    vscode.window.onDidChangeActiveTextEditor((editor) => {
      if (editor) heartbeat.handleActivity(editor.document, true);
    }),
  );

  // Send a heartbeat immediately on save
  context.subscriptions.push(
    vscode.workspace.onDidSaveTextDocument((doc) => {
      heartbeat.handleActivity(doc, true);
    }),
  );

  // Send a heartbeat when the VS Code window gains/loses focus
  context.subscriptions.push(
    vscode.window.onDidChangeWindowState(() => {
      const editor = vscode.window.activeTextEditor;
      if (editor) heartbeat.handleActivity(editor.document, true);
    }),
  );

  // Command: set API key
  context.subscriptions.push(
    vscode.commands.registerCommand('seismic.setApiKey', async () => {
      const key = await vscode.window.showInputBox({
        prompt: 'Enter your Seismic API key',
        placeHolder: 'Find your API key at seismic.icu/settings',
        password: true,
      });
      if (key) {
        await config.setApiKey(key);
        vscode.window.showInformationMessage('Seismic: API key saved!');
        statusBar.refresh();
      }
    }),
  );

  // Command: open dashboard
  context.subscriptions.push(
    vscode.commands.registerCommand('seismic.openDashboard', () => {
      vscode.env.openExternal(vscode.Uri.parse('https://seismic.icu/dashboard'));
    }),
  );

  // Command: enable tracking
  context.subscriptions.push(
    vscode.commands.registerCommand('seismic.enable', async () => {
      await vscode.workspace
        .getConfiguration('seismic')
        .update('enabled', true, vscode.ConfigurationTarget.Global);
      vscode.window.showInformationMessage('Seismic: Tracking enabled');
      statusBar.refresh();
    }),
  );

  // Command: disable tracking
  context.subscriptions.push(
    vscode.commands.registerCommand('seismic.disable', async () => {
      await vscode.workspace
        .getConfiguration('seismic')
        .update('enabled', false, vscode.ConfigurationTarget.Global);
      vscode.window.showInformationMessage('Seismic: Tracking disabled');
      statusBar.refresh();
    }),
  );

  statusBar.startUpdating();
  context.subscriptions.push({ dispose: () => statusBar.dispose() });

  // Retry any queued heartbeats every 5 minutes
  const flushInterval = setInterval(() => heartbeat.flushQueue(), 5 * 60 * 1000);
  context.subscriptions.push({ dispose: () => clearInterval(flushInterval) });
}

export function deactivate() {
  // VS Code disposes everything in context.subscriptions automatically
}
