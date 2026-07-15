import { Component, computed, inject, OnDestroy, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../core/api/api.service';
import { Heatmap } from '../../shared/components/heatmap/heatmap';
import { PieChart, PieSlice } from '../../shared/components/pie-chart/pie-chart';
import { ProjectBars, ProjectStat } from '../../shared/components/project-bars/project-bars';
import { TimelineChart, TimelineDay } from '../../shared/components/timeline-chart/timeline-chart';
import { GoalCard, GoalData } from '../../shared/components/goal-card/goal-card';
import { retry, Subscription, timer } from 'rxjs';

interface StatsSummary {
  totalSeconds: number;
  topLanguage: string | null;
  topProject: string | null;
  topOS: string | null;
  topEditor: string | null;
  dailyAverage: number;
  currentStreak: number;
  totalKeystrokes: number;
}

interface DashboardData {
  summary: StatsSummary;
  heatmap: HeatmapDay[];
  languages: { language: string; seconds: number }[];
  editors: { editor: string; seconds: number }[];
  os: { os: string; seconds: number }[];
  projects: ProjectStat[];
  timeline: TimelineDay[];
  goals: GoalData[];
}

interface HeatmapDay {
  date: string;
  seconds: number;
}

type RangeOption = 'today' | 'week' | 'month' | 'all';

const LABEL_OVERRIDES = new Map<string, string>([
  ['css', 'CSS'],
  ['html', 'HTML'],
  ['javascript', 'JavaScript'],
  ['typescript', 'TypeScript'],
  ['json', 'JSON'],
  ['jsx', 'JSX'],
  ['tsx', 'TSX'],
  ['yaml', 'YAML'],
  ['yml', 'YAML'],
  ['sql', 'SQL'],
  ['go', 'Go'],
  ['golang', 'Go'],
  ['gpt', 'Gpt'],
  ['macos', 'macOS'],
  ['ios', 'iOS'],
  ['linux', 'Linux'],
  ['windows', 'Windows'],
]);

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [FormsModule, Heatmap, PieChart, ProjectBars, TimelineChart, GoalCard],
  templateUrl: './dashboard.html',
})
export class Dashboard implements OnInit, OnDestroy {
  private api = inject(ApiService);
  private dashboardRequest?: Subscription;
  private latestRequestId = 0;

  range = signal<RangeOption>('week');
  loading = signal(true);
  errorMessage = signal<string | null>(null);

  stats = signal<StatsSummary | null>(null);
  heatmapData = signal<HeatmapDay[]>([]);
  languageData = signal<PieSlice[]>([]);
  editorData = signal<PieSlice[]>([]);
  osData = signal<PieSlice[]>([]);
  projectData = signal<ProjectStat[]>([]);
  timelineData = signal<TimelineDay[]>([]);
  goals = signal<GoalData[]>([]);

  ngOnInit() {
    this.loadAll();
  }

  ngOnDestroy() {
    this.dashboardRequest?.unsubscribe();
  }

  setRange(range: RangeOption) {
    if (this.range() === range) return;
    this.range.set(range);
    this.loadAll();
  }

  private loadAll() {
    const requestRange = this.range();
    const requestId = ++this.latestRequestId;

    this.dashboardRequest?.unsubscribe();
    this.loading.set(true);
    this.errorMessage.set(null);
    this.dashboardRequest = this.api
      .get<DashboardData>('/api/stats/dashboard', { range: requestRange })
      .pipe(retry({ count: 2, delay: (_, retryIndex) => timer(retryIndex * 500) }))
      .subscribe({
        next: (data) => {
          if (!this.isCurrentRequest(requestId, requestRange)) return;

          this.applyDashboardData(data);
          this.errorMessage.set(null);
          this.loading.set(false);
        },
        error: (err) => {
          if (!this.isCurrentRequest(requestId, requestRange)) return;

          console.error('Failed to load dashboard after retries:', err);
          this.clearDashboardData();
          this.errorMessage.set('Could not load dashboard data. Try again.');
          this.loading.set(false);
        },
      });
  }

  private isCurrentRequest(requestId: number, range: RangeOption): boolean {
    return requestId === this.latestRequestId && range === this.range();
  }

  private applyDashboardData(data: DashboardData): void {
    this.stats.set(data.summary);
    this.heatmapData.set(data.heatmap ?? []);
    this.languageData.set(
      (data.languages ?? []).map((d) => ({
        label: this.formatDisplayLabel(d.language),
        seconds: d.seconds,
      })),
    );
    this.editorData.set(
      (data.editors ?? []).map((d) => ({
        label: this.formatDisplayLabel(d.editor),
        seconds: d.seconds,
      })),
    );
    this.osData.set(
      (data.os ?? []).map((d) => ({
        label: this.formatDisplayLabel(d.os),
        seconds: d.seconds,
      })),
    );
    this.projectData.set(data.projects ?? []);
    this.timelineData.set(data.timeline ?? []);
    this.goals.set(data.goals ?? []);
  }

  private clearDashboardData(): void {
    this.stats.set(null);
    this.heatmapData.set([]);
    this.languageData.set([]);
    this.editorData.set([]);
    this.osData.set([]);
    this.projectData.set([]);
    this.timelineData.set([]);
    this.goals.set([]);
  }

  formatSeconds(seconds: number): string {
    if (seconds < 60) return '< 1m';
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
  }

  formatPreciseSeconds(seconds: number): string {
    if (seconds <= 0) return '0s';

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const remainingSeconds = seconds % 60;
    const parts: string[] = [];

    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    if (remainingSeconds > 0 || parts.length === 0) parts.push(`${remainingSeconds}s`);

    return parts.join(' ');
  }

  activitySummary(): string {
    const stats = this.stats();
    if (!stats) return '';

    const rangeLabel = this.summaryRangeLabel();
    const languages = this.languageData().map((item) => item.label);
    const editors = this.editorData().map((item) => item.label);

    let summary = `${rangeLabel}, you've logged ${this.formatPreciseSeconds(stats.totalSeconds)}`;

    const languageSummary = this.formatGroupedList(languages, 'language');
    if (languageSummary) summary += ` across ${languageSummary}`;

    const editorSummary = this.formatPlainList(editors.slice(0, 3));
    if (editorSummary) summary += ` using ${editorSummary}`;

    return summary;
  }

  formatDisplayLabel(value: string | null | undefined): string {
    if (!value) return '—';

    const trimmed = value.trim();
    if (!trimmed) return '—';

    const override = LABEL_OVERRIDES.get(trimmed.toLowerCase());
    if (override) return override;

    return trimmed
      .split(/([\s._-]+)/)
      .map((part) => {
        if (/^[\s._-]+$/.test(part)) return part;
        return part.charAt(0).toUpperCase() + part.slice(1);
      })
      .join('');
  }

  formatKeystrokes(count: number): string {
    if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M`;
    if (count >= 1_000) return `${(count / 1_000).toFixed(1)}k`;
    return count.toLocaleString();
  }

  isLastOdd(index: number, total: number): boolean {
    return total % 2 !== 0 && index === total - 1;
  }

  private summaryRangeLabel(): string {
    if (this.range() === 'today') return 'Today';
    if (this.range() === 'week') return 'This week';
    if (this.range() === 'month') return 'This month';
    return 'All time';
  }

  private formatGroupedList(items: string[], singularLabel: string): string {
    if (items.length <= 2) return this.formatPlainList(items);

    const visible = items.slice(0, 2).join(', ');
    const remaining = items.length - 2;
    const pluralLabel = remaining === 1 ? singularLabel : `${singularLabel}s`;

    return `${visible} (& ${remaining} other ${pluralLabel})`;
  }

  private formatPlainList(items: string[]): string {
    if (items.length === 0) return '';
    if (items.length === 1) return items[0];
    if (items.length === 2) return `${items[0]} and ${items[1]}`;

    return `${items.slice(0, -1).join(', ')}, and ${items[items.length - 1]}`;
  }

  limitedProjectData = computed(() => this.projectData().slice(0, 10));
}
