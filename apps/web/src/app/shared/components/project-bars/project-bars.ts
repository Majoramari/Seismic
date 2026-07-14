import { Component, input, computed } from '@angular/core';

export interface ProjectStat {
  project: string;
  seconds: number;
}

interface DisplayBar extends ProjectStat {
  widthPercent: number;
  formattedTime: string;
  label: string;
}

@Component({
  selector: 'app-project-bars',
  standalone: true,
  templateUrl: './project-bars.html',
})
export class ProjectBars {
  data = input.required<ProjectStat[]>();

  bars = computed<DisplayBar[]>(() => {
    const items = this.data();
    if (items.length === 0) return [];

    const max = Math.max(...items.map((i) => i.seconds));

    return items.map((item) => ({
      ...item,
      widthPercent: max > 0 ? (item.seconds / max) * 100 : 0,
      formattedTime: this.formatSeconds(item.seconds),
      label: item.project.trim() || 'Unnamed project',
    }));
  });

  private formatSeconds(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  }
}
