import { Component, input, computed } from '@angular/core';
import { CommonModule } from '@angular/common';

export interface TimelineDay {
  date: string;
  seconds: number;
}

interface DisplayBar extends TimelineDay {
  heightPercent: number;
  label: string;
}

@Component({
  selector: 'app-timeline-chart',
  standalone: true,
  imports: [CommonModule],
  templateUrl: './timeline-chart.html',
})
export class TimelineChart {
  data = input.required<TimelineDay[]>();

  maxHours = computed(() => {
    const max = Math.max(...this.data().map((d) => d.seconds), 3600);
    return Math.ceil(max / 3600);
  });

  bars = computed<DisplayBar[]>(() => {
    const maxSeconds = this.maxHours() * 3600;
    return this.data().map((d) => ({
      ...d,
      heightPercent: maxSeconds > 0 ? (d.seconds / maxSeconds) * 100 : 0,
      label: new Date(d.date).toLocaleDateString('en-US', { month: 'short', day: 'numeric' }),
    }));
  });

  yAxisLabels = computed(() => {
    const max = this.maxHours();
    const step = Math.max(1, Math.round(max / 4));
    const labels: number[] = [];
    for (let i = 0; i <= max; i += step) {
      labels.push(i);
    }
    return labels.reverse();
  });
}
