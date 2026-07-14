import { Component, computed, signal } from '@angular/core';
import { RouterLink } from '@angular/router';
import {
  BarChart3,
  BookOpen,
  Clock,
  Code2,
  Download,
  FolderGit2,
  Github,
  LucideAngularModule,
  Monitor,
  Shield,
  Target,
  Terminal,
  Trophy,
  WifiOff,
} from 'lucide-angular';

type PreviewTab = 'dashboard' | 'projects' | 'goals' | 'settings';

@Component({
  selector: 'app-landing',
  standalone: true,
  imports: [RouterLink, LucideAngularModule],
  templateUrl: './landing.html',
})
export class Landing {
  readonly ClockIcon = Clock;
  readonly CodeIcon = Code2;
  readonly ProjectIcon = FolderGit2;
  readonly GoalIcon = Target;
  readonly TrophyIcon = Trophy;
  readonly OfflineIcon = WifiOff;
  readonly ShieldIcon = Shield;
  readonly DownloadIcon = Download;
  readonly MonitorIcon = Monitor;
  readonly TerminalIcon = Terminal;
  readonly BookIcon = BookOpen;
  readonly GithubIcon = Github;
  readonly ChartIcon = BarChart3;

  readonly previewTabs: { id: PreviewTab; label: string }[] = [
    { id: 'dashboard', label: 'Dashboard' },
    { id: 'projects', label: 'Projects' },
    { id: 'goals', label: 'Goals' },
    { id: 'settings', label: 'Settings' },
  ];

  readonly activePreviewTab = signal<PreviewTab>('dashboard');
  readonly easterEggActive = signal(false);
  readonly preview = computed(() => this.previewContent[this.activePreviewTab()]);
  private goalClickCount = 0;
  private lastGoalClickAt = 0;

  private readonly previewContent: Record<
    PreviewTab,
    {
      eyebrow: string;
      title: string;
      cards: { label: string; value: string }[];
    }
  > = {
    dashboard: {
      eyebrow: 'All time, before it slips away',
      title: 'Keep Track of Your Coding Time',
      cards: [
        { label: 'Total time', value: '843h 31m' },
        { label: 'Top project', value: 'Seismic' },
        { label: 'Top language', value: 'TypeScript' },
        { label: 'Top editor', value: 'VS Code' },
      ],
    },
    projects: {
      eyebrow: 'Project history, grouped automatically',
      title: 'Know Where Your Focus Goes',
      cards: [
        { label: 'Active projects', value: '12' },
        { label: 'Top repo', value: 'Seismic' },
        { label: 'Last commit', value: '2h ago' },
        { label: 'This week', value: '18h 42m' },
      ],
    },
    goals: {
      eyebrow: 'Progress without manual timers',
      title: 'Goals That Update Themselves',
      cards: [
        { label: 'Monthly goal', value: '72%' },
        { label: 'Current streak', value: '9 days' },
        { label: 'Badges earned', value: '14' },
        { label: 'Reminder status', value: 'On' },
      ],
    },
    settings: {
      eyebrow: 'Privacy and editor controls',
      title: 'Tune Tracking To Your Workflow',
      cards: [
        { label: 'Profile visibility', value: 'Public' },
        { label: 'Hidden projects', value: '3' },
        { label: 'Project naming', value: 'Git root' },
        { label: 'API key', value: 'Ready' },
      ],
    },
  };

  setPreviewTab(tab: PreviewTab): void {
    this.activePreviewTab.set(tab);
    if (tab === 'goals') {
      this.recordGoalClick();
    }
  }

  private recordGoalClick(): void {
    const now = Date.now();
    this.goalClickCount = now - this.lastGoalClickAt < 900 ? this.goalClickCount + 1 : 1;
    this.lastGoalClickAt = now;

    if (this.goalClickCount >= 5) {
      this.easterEggActive.set(true);
      this.goalClickCount = 0;
    }
  }
}
