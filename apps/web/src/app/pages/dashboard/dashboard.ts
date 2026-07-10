import { Component, effect, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../core/api/api.service';
import { Heatmap } from '../../shared/components/heatmap/heatmap';
import { PieChart, PieSlice } from '../../shared/components/pie-chart/pie-chart';
import { ProjectBars, ProjectStat } from '../../shared/components/project-bars/project-bars';
import { TimelineChart, TimelineDay } from '../../shared/components/timeline-chart/timeline-chart';
import { GoalCard, GoalData } from '../../shared/components/goal-card/goal-card';

interface StatsSummary {
  totalSeconds: number;
  topLanguage: string | null;
  topProject: string | null;
  topOS: string | null;
  topEditor: string | null;
  dailyAverage: number;
  currentStreak: number;
}

interface DashboardData {
  summary: StatsSummary;
  heatmap: HeatmapDay[];
  languages: { language: string; seconds: number }[];
  editors: { editor: string; seconds: number }[];
  os: { os: string; seconds: number }[];
  projects: ProjectStat[];
  timeline: TimelineDay[];
}

interface HeatmapDay {
  date: string;
  seconds: number;
}

type RangeOption = 'today' | 'week' | 'month' | 'all';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [FormsModule, Heatmap, PieChart, ProjectBars, TimelineChart, GoalCard],
  templateUrl: './dashboard.html',
})
export class Dashboard implements OnInit {
  private api = inject(ApiService);

  range = signal<RangeOption>('week');
  loading = signal(true);

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
    this.loadGoals();
  }

  setRange(range: RangeOption) {
    this.range.set(range);
    this.loadAll();
  }

  private loadGoals() {
    this.api.get<GoalData[]>('/api/goals').subscribe({
      next: (data) => this.goals.set(data ?? []),
      error: () => {},
    });
  }

  private loadAll() {
    this.loading.set(true);
    this.api.get<DashboardData>('/api/stats/dashboard', { range: this.range() }).subscribe({
      next: (data) => {
        this.stats.set(data.summary);
        this.heatmapData.set(data.heatmap ?? []);
        this.languageData.set(
          (data.languages ?? []).map((d) => ({ label: d.language, seconds: d.seconds })),
        );
        this.editorData.set(
          (data.editors ?? []).map((d) => ({ label: d.editor, seconds: d.seconds })),
        );
        this.osData.set((data.os ?? []).map((d) => ({ label: d.os, seconds: d.seconds })));
        this.projectData.set(data.projects ?? []);
        this.timelineData.set(data.timeline ?? []);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }

  formatSeconds(seconds: number): string {
    if (seconds < 60) return '< 1m';
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
  }

  isLastOdd(index: number, total: number): boolean {
    return total % 2 !== 0 && index === total - 1;
  }
}
