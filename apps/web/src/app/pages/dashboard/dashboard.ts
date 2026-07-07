import { Component, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../core/api/api.service';
import { Heatmap } from '../../shared/components/heatmap/heatmap';
import { PieChart, PieSlice } from '../../shared/components/pie-chart/pie-chart';
import { ProjectBars, ProjectStat } from '../../shared/components/project-bars/project-bars';
import { TimelineChart, TimelineDay } from '../../shared/components/timeline-chart/timeline-chart';

interface StatsSummary {
  totalSeconds: number;
  topLanguage: string | null;
  topProject: string | null;
  dailyAverage: number;
  currentStreak: number;
}

interface HeatmapDay {
  date: string;
  seconds: number;
}

type RangeOption = 'today' | 'week' | 'month' | 'all';

@Component({
  selector: 'app-dashboard',
  standalone: true,
  imports: [FormsModule, Heatmap, PieChart, ProjectBars, TimelineChart],
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
  projectData = signal<ProjectStat[]>([]);
  timelineData = signal<TimelineDay[]>([]);

  ngOnInit() {
    this.loadAll();
  }

  setRange(range: RangeOption) {
    this.range.set(range);
    this.loadAll();
  }

  private loadAll() {
    this.loading.set(true);
    const range = this.range();

    this.api.get<StatsSummary>('/api/stats/summary', { range }).subscribe({
      next: (data) => {
        this.stats.set(data);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });

    this.api.get<HeatmapDay[]>('/api/stats/heatmap').subscribe({
      next: (data) => this.heatmapData.set(data ?? []),
      error: () => {},
    });

    this.api
      .get<{ language: string; seconds: number }[]>('/api/stats/languages', { range })
      .subscribe({
        next: (data) =>
          this.languageData.set(
            (data ?? []).map((d) => ({ label: d.language, seconds: d.seconds })),
          ),
        error: () => {},
      });

    this.api.get<{ editor: string; seconds: number }[]>('/api/stats/editors', { range }).subscribe({
      next: (data) =>
        this.editorData.set((data ?? []).map((d) => ({ label: d.editor, seconds: d.seconds }))),
      error: () => {},
    });

    this.api.get<ProjectStat[]>('/api/stats/projects', { range }).subscribe({
      next: (data) => this.projectData.set(data ?? []),
      error: () => {},
    });

    this.api.get<TimelineDay[]>('/api/stats/timeline').subscribe({
      next: (data) => this.timelineData.set(data ?? []),
      error: () => {},
    });
  }

  formatSeconds(seconds: number): string {
    if (seconds < 60) return '< 1m';
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return hours > 0 ? `${hours}h ${minutes}m` : `${minutes}m`;
  }
}
