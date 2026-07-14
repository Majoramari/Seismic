import * as vscode from 'vscode';
import * as config from './config';
import { HeartbeatService } from './heartbeat';
import { ProjectSyncService } from './projectSync';
import { StatusBarManager } from './statusbar';

/**
 * Entry point for the Seismic extension. VS Code calls
 * activate() once when the extension starts, and deactivate()
 * when it shuts down.
 */
export function activate(context: vscode.ExtensionContext) {
    config.configureDefaultApiUrl(context.extensionMode === vscode.ExtensionMode.Development);

    const heartbeat = new HeartbeatService();
    const projectSync = new ProjectSyncService();
    const statusBar = new StatusBarManager();

    void config.refreshEditorSettings().finally(() => projectSync.syncWorkspaceFolders(true));

    // Every one of these just records "something happened" — none of
    // them send a network request directly. heartbeat.start() below is
    // the only thing that actually fires a heartbeat, on a fixed
    // 30-second interval, so switching tabs or saving rapidly can't
    // cause a burst of requests anymore.

    // Typing: also tally characters inserted toward the keystroke count.
    context.subscriptions.push(
        vscode.workspace.onDidChangeTextDocument((e) => {
            const realKeystrokes = e.contentChanges.filter(
                (change) => change.text.length === 1 && change.rangeLength === 0,
            ).length;
            heartbeat.recordKeystrokes(realKeystrokes);
            heartbeat.noteActivity(e.document);
        }),
    );

    // Switching files
    context.subscriptions.push(
        vscode.window.onDidChangeActiveTextEditor((editor) => {
            if (editor) heartbeat.noteActivity(editor.document);
        }),
    );

    // Saving
    context.subscriptions.push(
        vscode.workspace.onDidSaveTextDocument((doc) => {
            heartbeat.noteActivity(doc);
            void projectSync.syncDocument(doc, true);
        }),
    );

    context.subscriptions.push(
        vscode.workspace.onDidChangeWorkspaceFolders((event) => {
            for (const folder of event.added) {
                void projectSync.syncWorkspaceFolder(folder, true);
            }
        }),
    );

    // Window gaining focus
    context.subscriptions.push(
        vscode.window.onDidChangeWindowState((state) => {
            if (!state.focused) return; // only care about gaining focus, not losing it
            const editor = vscode.window.activeTextEditor;
            if (editor) heartbeat.noteActivity(editor.document);
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
                await config.refreshEditorSettings();
                vscode.window.showInformationMessage('Seismic: API key saved!');
                statusBar.refresh();
                void projectSync.syncWorkspaceFolders(true);
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

    // Start the periodic heartbeat tick (see heartbeat.ts) — this replaces
    // the old immediate-send-on-save/switch behavior.
    heartbeat.start();
    context.subscriptions.push({ dispose: () => heartbeat.dispose() });

    // Retry any queued (failed) heartbeats every 5 minutes
    const flushInterval = setInterval(() => heartbeat.flushQueue(), 5 * 60 * 1000);
    context.subscriptions.push({ dispose: () => clearInterval(flushInterval) });
}

export function deactivate() {
    // VS Code disposes everything in context.subscriptions automatically
}
