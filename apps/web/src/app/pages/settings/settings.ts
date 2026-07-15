import { Component, inject, signal, OnInit, computed } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../core/api/api.service';
import { ToastService } from '../../core/toast/toast.service';
import { CommonModule } from '@angular/common';
import { Code2, KeyRound, LucideAngularModule, Shield, Target, User, Award, Download } from 'lucide-angular';
import { badgeInfo } from '../../shared/badges';

interface ApiKeyResponse {
  apiKey: string;
}

interface PrivacySettings {
  hideProjects: boolean;
  hideTime: boolean;
  hideLanguages: boolean;
  hideOS: boolean;
  hideEditor: boolean;
  hideLeaderboard: boolean;
  profilePublic: boolean;
}

interface EditorSettings {
  useGitRootProjectName: boolean;
}

interface GoalData {
  id: string;
  scope: string;
  scopeValue: string | null;
  period: string;
  targetSeconds: number;
  remindersEnabled: boolean;
}

interface CurrentUserResponse {
  email: string;
}

interface ImportProgress {
  status: 'idle' | 'processing' | 'finalizing' | 'completed' | 'failed';
  imported: number;
  total: number;
  error?: string;
}

interface SettingsBadge {
  type: string;
  earnedAt: string;
  hidden: boolean;
}

type SettingsSection = 'apikey' | 'privacy' | 'editors' | 'goals' | 'badges' | 'account' | 'imports';
type AccountSubTab = 'reset' | 'email' | 'delete';

@Component({
  selector: 'app-settings',
  standalone: true,
  imports: [FormsModule, CommonModule, LucideAngularModule],
  templateUrl: './settings.html',
})
export class Settings implements OnInit {
  private api = inject(ApiService);
  private toast = inject(ToastService);

  readonly KeyIcon = KeyRound;
  readonly ShieldIcon = Shield;
  readonly TargetIcon = Target;
  readonly UserIcon = User;
  readonly AwardIcon = Award;
  readonly DownloadIcon = Download;
  readonly CodeIcon = Code2;

  activeSection = signal<SettingsSection>('apikey');
  accountSubTab = signal<AccountSubTab>('email');

  apiKey = signal('');
  apiKeyVisible = signal(false);
  regenerating = signal(false);

  newEmail = signal('');
  currentEmail = signal('');
  changingEmail = signal(false);

  privacy = signal<PrivacySettings | null>(null);
  privacyDirty = signal(false);
  savingPrivacy = signal(false);

  editorSettings = signal<EditorSettings | null>(null);
  savingEditorSettings = signal(false);

  showDeleteConfirm = signal(false);
  deleteConfirmText = signal('');

  // Goals state
  goals = signal<GoalData[]>([]);
  languages = signal<string[]>([]);
  projects = signal<string[]>([]);
  showGoalModal = signal(false);
  editingGoalId = signal<string | null>(null);
  savingGoal = signal(false);

  goalAmount = 30;
  goalUnit: 'minutes' | 'hours' = 'minutes';
  goalPeriod: 'daily' | 'weekly' | 'monthly' = 'daily';
  goalLanguage = '';
  goalProject = '';
  goalReminders = false;

  wakaTimeApiKey = '';
  importing = signal(false);
  importProgress = signal<ImportProgress | null>(null);
  importProvider: 'wakatime' | 'hackatime' = 'wakatime';
  selectedFile: File | null = null;
  private importPollInterval: ReturnType<typeof setInterval> | null = null;

  badges = signal<SettingsBadge[]>([]);
  badgesLoading = signal(true);
  sortedBadges = computed(() =>
    [...this.badges()].sort((a, b) => this.badgeLabel(a.type).localeCompare(this.badgeLabel(b.type))),
  );

  ngOnInit() {
    this.loadApiKey();
    this.loadPrivacy();
    this.loadEditorSettings();
    this.loadGoals();
    this.loadFilters();
    this.loadCurrentUser();
    this.loadBadges();
  }

  private loadCurrentUser() {
    this.api.get<CurrentUserResponse>('/api/auth/me').subscribe({
      next: (data) => this.currentEmail.set(data.email),
      error: () => {},
    });
  }

  setSection(section: SettingsSection) {
    this.activeSection.set(section);
  }

  setAccountSubTab(tab: AccountSubTab) {
    this.accountSubTab.set(tab);
  }

  // --- API Key ---
  private loadApiKey() {
    this.api.get<ApiKeyResponse>('/api/auth/apikey').subscribe({
      next: (data) => this.apiKey.set(data.apiKey),
      error: () => this.toast.error('Failed to load API key'),
    });
  }

  toggleApiKeyVisible() {
    this.apiKeyVisible.set(!this.apiKeyVisible());
  }

  copyApiKey() {
    navigator.clipboard.writeText(this.apiKey());
    this.toast.success('API key copied to clipboard');
  }

  regenerateApiKey() {
    this.regenerating.set(true);
    this.api.post<ApiKeyResponse>('/api/auth/apikey/regenerate', {}).subscribe({
      next: (data) => {
        this.apiKey.set(data.apiKey);
        this.regenerating.set(false);
        this.toast.success('API key regenerated. Update it in your editors.');
      },
      error: () => {
        this.regenerating.set(false);
        this.toast.error('Failed to regenerate API key');
      },
    });
  }

  // --- Privacy ---
  private loadPrivacy() {
    this.api.get<PrivacySettings>('/api/settings/privacy').subscribe({
      next: (data) => this.privacy.set(data),
      error: () => this.toast.error('Failed to load privacy settings'),
    });
  }

  toggleLocal(key: keyof PrivacySettings, value: boolean) {
    const current = this.privacy();
    if (!current) return;
    this.privacy.set({ ...current, [key]: value });
    this.privacyDirty.set(true);
  }

  savePrivacy() {
    const current = this.privacy();
    if (!current) return;

    this.savingPrivacy.set(true);
    this.api.post('/api/settings/privacy', current).subscribe({
      next: () => {
        this.savingPrivacy.set(false);
        this.privacyDirty.set(false);
        this.toast.success('Privacy settings saved');
      },
      error: () => {
        this.savingPrivacy.set(false);
        this.toast.error('Failed to save settings');
      },
    });
  }

  private loadEditorSettings() {
    this.api.get<EditorSettings>('/api/settings/editor').subscribe({
      next: (data) => this.editorSettings.set(data),
      error: () => this.toast.error('Failed to load editor settings'),
    });
  }

  toggleEditorSetting(key: keyof EditorSettings, value: boolean) {
    const current = this.editorSettings();
    if (!current) return;

    const next = { ...current, [key]: value };
    this.editorSettings.set(next);
    this.savingEditorSettings.set(true);
    this.api.post('/api/settings/editor', next).subscribe({
      next: () => {
        this.savingEditorSettings.set(false);
        this.toast.success('Editor settings saved');
      },
      error: () => {
        this.editorSettings.set(current);
        this.savingEditorSettings.set(false);
        this.toast.error('Failed to save editor settings');
      },
    });
  }

  // --- Goals ---
  private loadGoals() {
    this.api.get<GoalData[]>('/api/goals').subscribe({
      next: (data) => this.goals.set(data ?? []),
      error: () => {},
    });
  }

  sortedGoals = computed(() => [...this.goals()].sort((a, b) => a.targetSeconds - b.targetSeconds));

  private loadFilters() {
    this.api.get<string[]>('/api/filters/languages').subscribe({
      next: (data) => this.languages.set(data ?? []),
      error: () => {},
    });
    this.api.get<string[]>('/api/filters/projects').subscribe({
      next: (data) => this.projects.set(data ?? []),
      error: () => {},
    });
  }

  openNewGoalModal() {
    this.editingGoalId.set(null);
    this.goalAmount = 30;
    this.goalUnit = 'minutes';
    this.goalPeriod = 'daily';
    this.goalLanguage = '';
    this.goalProject = '';
    this.goalReminders = false;
    this.showGoalModal.set(true);
  }

  openEditGoalModal(goal: GoalData) {
    this.editingGoalId.set(goal.id);
    const isHours = goal.targetSeconds >= 3600 && goal.targetSeconds % 3600 === 0;
    this.goalUnit = isHours ? 'hours' : 'minutes';
    this.goalAmount = isHours ? goal.targetSeconds / 3600 : goal.targetSeconds / 60;
    this.goalPeriod = goal.period as 'daily' | 'weekly';
    this.goalLanguage = goal.scope === 'language' ? (goal.scopeValue ?? '') : '';
    this.goalProject = goal.scope === 'project' ? (goal.scopeValue ?? '') : '';
    this.goalReminders = goal.remindersEnabled;
    this.showGoalModal.set(true);
  }

  closeGoalModal() {
    this.showGoalModal.set(false);
  }

  setQuickPick(minutes: number) {
    if (minutes >= 60) {
      this.goalUnit = 'hours';
      this.goalAmount = minutes / 60;
    } else {
      this.goalUnit = 'minutes';
      this.goalAmount = minutes;
    }
  }

  isQuickPickActive(minutes: number): boolean {
    const currentMinutes = this.goalUnit === 'hours' ? this.goalAmount * 60 : this.goalAmount;
    return currentMinutes === minutes;
  }

  saveGoal() {
    const targetSeconds = this.goalUnit === 'hours' ? this.goalAmount * 3600 : this.goalAmount * 60;

    let scope = 'overall';
    let scopeValue: string | null = null;
    if (this.goalLanguage) {
      scope = 'language';
      scopeValue = this.goalLanguage;
    } else if (this.goalProject) {
      scope = 'project';
      scopeValue = this.goalProject;
    }

    const body = {
      scope,
      scopeValue,
      period: this.goalPeriod,
      targetSeconds: Math.round(targetSeconds),
      remindersEnabled: this.goalReminders,
    };

    this.savingGoal.set(true);
    const editingId = this.editingGoalId();

    const request = editingId
      ? this.api.put<GoalData>(`/api/goals/${editingId}`, body)
      : this.api.post<GoalData>('/api/goals', body);

    request.subscribe({
      next: (goal) => {
        if (editingId) {
          this.goals.set(this.goals().map((g) => (g.id === editingId ? goal : g)));
        } else {
          this.goals.set([...this.goals(), goal]);
        }
        this.savingGoal.set(false);
        this.showGoalModal.set(false);
        this.toast.success(editingId ? 'Goal updated' : 'Goal created');
      },
      error: () => {
        this.savingGoal.set(false);
        this.toast.error('Failed to save goal');
      },
    });
  }

  deleteGoal(id: string) {
    const confirmed = confirm('Delete this goal?');
    if (!confirmed) return;

    this.api.delete(`/api/goals/${id}`).subscribe({
      next: () => {
        this.goals.set(this.goals().filter((g) => g.id !== id));
        this.toast.success('Goal deleted');
      },
      error: () => this.toast.error('Failed to delete goal'),
    });
  }

  formatGoalTime(goal: GoalData): string {
    const hours = Math.floor(goal.targetSeconds / 3600);
    const minutes = Math.floor((goal.targetSeconds % 3600) / 60);
    if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`;
    if (hours > 0) return `${hours}h`;
    return `${minutes}m`;
  }

  goalLabel(goal: GoalData): string {
    if (goal.scope === 'language') return `Language: ${goal.scopeValue}`;
    if (goal.scope === 'project') return `Project: ${goal.scopeValue}`;
    return 'All programming activity';
  }

  onLanguageChange(value: string) {
    this.goalLanguage = value;
    if (value) this.goalProject = '';
  }

  onProjectChange(value: string) {
    this.goalProject = value;
    if (value) this.goalLanguage = '';
  }

  // --- Account ---
  resetTimers() {
    const confirmed = confirm(
      'This will permanently delete all your tracked coding time. This cannot be undone. Continue?',
    );
    if (!confirmed) return;

    this.api.post('/api/settings/reset-timers', {}).subscribe({
      next: () => this.toast.success('All coding stats have been reset'),
      error: () => this.toast.error('Failed to reset timers'),
    });
  }

  confirmDelete() {
    if (this.deleteConfirmText() !== 'delete') return;

    this.api.post('/api/settings/account', {}).subscribe({
      next: () => {
        localStorage.clear();
        window.location.href = '/login';
      },
      error: () => this.toast.error('Failed to delete account'),
    });
  }

  requestEmailChange() {
    if (!this.newEmail()) return;

    this.changingEmail.set(true);
    this.api.post('/api/auth/change-email', { newEmail: this.newEmail() }).subscribe({
      next: () => {
        this.changingEmail.set(false);
        this.toast.success('Check your new email to confirm the change');
        this.newEmail.set('');
      },
      error: (err) => {
        this.changingEmail.set(false);
        this.toast.error(err.error?.message || 'Failed to request email change');
      },
    });
  }

  onFileSelected(event: Event) {
    const input = event.target as HTMLInputElement;
    this.selectedFile = input.files?.[0] ?? null;
  }

  startWakaTimeImport() {
    if (!this.wakaTimeApiKey.trim()) {
      this.toast.error('Enter your WakaTime API key');
      return;
    }
    this.importing.set(true);
    this.api.post('/api/import/wakatime', { apiKey: this.wakaTimeApiKey.trim() }).subscribe({
      next: () => this.toast.success('Import started — check your dashboard in a few minutes'),
      error: (err) => {
        this.importing.set(false);
        this.toast.error(err.error?.message ?? 'Failed to start import');
      },
    });
  }

  importFromFile() {
    if (!this.selectedFile) {
      this.toast.error('Choose a file first');
      return;
    }

    const formData = new FormData();
    formData.append('file', this.selectedFile);

    this.importing.set(true);
    this.api.postFormData('/api/import/file', formData).subscribe({
      next: () => {
        this.pollImportStatus();
      },
      error: (err) => {
        this.importing.set(false);
        this.toast.error(err.error?.message ?? 'Import failed');
      },
    });
  }

  private pollImportStatus() {
    this.importPollInterval = setInterval(() => {
      this.api
        .get<ImportProgress>('/api/import/status')
        .subscribe({
          next: (status) => {
            this.importProgress.set(status);

            if (status.status === 'completed') {
              this.stopPolling();
              this.importing.set(false);
              this.toast.success(`Imported ${status.imported} heartbeats`);
            } else if (status.status === 'failed') {
              this.stopPolling();
              this.importing.set(false);
              this.toast.error(status.error ?? 'Import failed');
            }
          },
          error: () => this.stopPolling(),
        });
    }, 1500);
  }

  private stopPolling() {
    if (this.importPollInterval) {
      clearInterval(this.importPollInterval);
      this.importPollInterval = null;
    }
  }

  importProgressText(progress: ImportProgress): string {
    if (progress.status === 'finalizing') {
      return `Rebuilding stats from ${progress.imported} imported heartbeats...`;
    }
    return `${progress.imported} / ${progress.total} heartbeats imported`;
  }

  badgeLabel(type: string): string {
    return badgeInfo(type).label;
  }

  badgeDescription(type: string): string {
    return badgeInfo(type).description;
  }

  badgeColor(type: string): string {
    return badgeInfo(type).color;
  }

  settingsBadgeColor(badge: SettingsBadge): string {
    return badge.hidden ? '#9ca3af' : this.badgeColor(badge.type);
  }

  isBadgeHidden(type: string): boolean {
    return this.badges().some((badge) => badge.type === type && badge.hidden);
  }

  private loadBadges() {
    this.badgesLoading.set(true);
    this.api.get<SettingsBadge[]>('/api/settings/badges').subscribe({
      next: (data) => {
        this.badges.set(data ?? []);
        this.badgesLoading.set(false);
      },
      error: () => {
        this.badgesLoading.set(false);
      },
    });
  }

  toggleBadgeVisibility(type: string, hidden: boolean) {
    this.api.post('/api/settings/badge-visibility', { badgeType: type, hidden }).subscribe({
      next: () => {
        this.badges.set(
          this.badges().map((badge) => (badge.type === type ? { ...badge, hidden } : badge)),
        );
        this.toast.success('Badge visibility updated');
      },
      error: () => this.toast.error('Failed to update badge visibility'),
    });
  }
}
