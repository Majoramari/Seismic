import { CommonModule } from '@angular/common';
import { Component, ElementRef, computed, effect, input, viewChild } from '@angular/core';

export interface TimelineProject {
  project: string;
  seconds: number;
}

export interface TimelineDay {
  date: string;
  seconds: number;
  projects?: TimelineProject[];
}

interface DisplayProject extends TimelineProject {
  percentage: number;
  color: string;
  label: string;
}

interface DisplayBar extends TimelineDay {
  heightPercent: number;
  label: string;
  fullLabel: string;
  formattedTime: string;
  displayProjects: DisplayProject[];
  hoverProjects: DisplayProject[];
}

@Component({
  selector: 'app-timeline-chart',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './timeline-chart.html',
})
export class TimelineChart {
  data = input.required<TimelineDay[]>();

  private timelineBars = viewChild<ElementRef<HTMLDivElement>>('timelineBars');

  maxHours = computed(() => {
    const maximum = Math.max(...this.data().map((day) => day.seconds), 3600);

    return Math.ceil(maximum / 3600);
  });

  bars = computed<DisplayBar[]>(() => {
    const maxSeconds = this.maxHours() * 3600;

    return this.data().map((day) => {
      const dayProjects = this.toDisplayProjects(day.projects ?? [], day.seconds);

      return {
        ...day,
        heightPercent: maxSeconds > 0 ? (day.seconds / maxSeconds) * 100 : 0,
        label: this.formatDate(day.date, {
          month: 'short',
          day: 'numeric',
        }),
        fullLabel: this.formatDate(day.date, {
          weekday: 'long',
          month: 'long',
          day: 'numeric',
          year: 'numeric',
        }),
        formattedTime: this.formatSeconds(day.seconds),
        displayProjects: dayProjects,
        hoverProjects: dayProjects,
      };
    });
  });

  yAxisLabels = computed(() => {
    const maximum = this.maxHours();
    const step = Math.max(1, Math.ceil(maximum / 4));
    const labels: number[] = [];

    for (let value = 0; value <= maximum; value += step) {
      labels.push(value);
    }

    if (labels.at(-1) !== maximum) {
      labels.push(maximum);
    }

    return labels.reverse();
  });

  constructor() {
    effect(() => {
      this.bars();
      const element = this.timelineBars()?.nativeElement;

      if (!element) return;

      window.setTimeout(() => {
        element.scrollLeft = element.scrollWidth - element.clientWidth;
      });
    });
  }

  private toDisplayProjects(projects: TimelineProject[], total: number): DisplayProject[] {
    return projects
      .filter((project) => project.seconds > 0)
      .sort((first, second) => second.seconds - first.seconds)
      .map((project) => ({
        ...project,
        percentage: total > 0 ? (project.seconds / total) * 100 : 0,
        color: this.projectColor(project.project),
        label: project.project.trim() || 'Unnamed project',
      }));
  }

  private projectColor(projectName: string): string {
    let hash = 0;

    for (const character of projectName) {
      hash = character.charCodeAt(0) + ((hash << 5) - hash);
    }

    return `hsl(${Math.abs(hash) % 360} 72% 60%)`;
  }

  private formatDate(date: string, options: Intl.DateTimeFormatOptions): string {
    return new Date(`${date.slice(0, 10)}T00:00:00`).toLocaleDateString('en-US', options);
  }

  private formatSeconds(seconds: number): string {
    if (seconds <= 0) return 'No activity';

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (hours > 0) return `${hours}h ${minutes}m`;
    if (minutes > 0) return `${minutes}m`;

    return '< 1m';
  }
}
