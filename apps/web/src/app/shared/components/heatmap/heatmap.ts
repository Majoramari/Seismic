import { Component, input, computed } from '@angular/core';

interface HeatmapDay {
  date: string;
  seconds: number;
}

interface DisplayDay {
  date: string;
  seconds: number;
  level: number; // 0-4
  month: string;
}

@Component({
  selector: 'app-heatmap',
  standalone: true,
  templateUrl: './heatmap.html',
})
export class Heatmap {
  data = input.required<HeatmapDay[]>();

  weeks = computed(() => {
    const dataMap = new Map(this.data().map((d) => [d.date, d.seconds]));
    const days: DisplayDay[] = [];

    const today = new Date();
    const start = new Date(today);
    start.setDate(start.getDate() - 364);

    // Align start to the previous Sunday so weeks line up in columns
    const dayOfWeek = start.getDay();
    start.setDate(start.getDate() - dayOfWeek);

    const cursor = new Date(start);
    while (cursor <= today) {
      const dateStr = cursor.toISOString().split('T')[0];
      const seconds = dataMap.get(dateStr) ?? 0;
      days.push({
        date: dateStr,
        seconds,
        level: this.getLevel(seconds),
        month: cursor.toLocaleDateString('en-US', { month: 'short' }),
      });
      cursor.setDate(cursor.getDate() + 1);
    }

    // Chunk into weeks of 7
    const result: DisplayDay[][] = [];
    for (let i = 0; i < days.length; i += 7) {
      result.push(days.slice(i, i + 7));
    }
    return result;
  });

  private getLevel(seconds: number): number {
    if (seconds === 0) return 0;
    if (seconds < 1800) return 1;
    if (seconds < 3600) return 2;
    if (seconds < 7200) return 3;
    return 4;
  }

  formatTooltip(day: DisplayDay): string {
    const hours = Math.floor(day.seconds / 3600);
    const minutes = Math.floor((day.seconds % 3600) / 60);
    const timeStr = day.seconds === 0 ? 'No activity' : `${hours}h ${minutes}m`;
    const dateStr = new Date(day.date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
    return `${timeStr} on ${dateStr}`;
  }
}
