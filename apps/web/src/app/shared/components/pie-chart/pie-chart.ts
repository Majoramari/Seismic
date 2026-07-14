import { Component, input, computed, signal } from '@angular/core';
import { getLanguageColor } from './language-colors';

export interface PieSlice {
  label: string;
  seconds: number;
}

interface ComputedSlice extends PieSlice {
  id: string;
  percentage: number;
  color: string;
  pathD: string;
}

const GENERIC_COLORS = [
  '#e8c547',
  '#3b82f6',
  '#ec4899',
  '#a855f7',
  '#ef4444',
  '#f97316',
  '#0ea5e9',
  '#22c55e',
  '#8b5cf6',
  '#14b8a6',
];

@Component({
  selector: 'app-pie-chart',
  standalone: true,
  templateUrl: './pie-chart.html',
})
export class PieChart {
  data = input.required<PieSlice[]>();
  emptyLabel = input('No data yet');
  useLanguageColors = input(false);
  legendLimit = input<number | null>(null);
  hoveredSlice = signal<ComputedSlice | null>(null);

  displayData = computed<PieSlice[]>(() => {
    const data = this.data();
    const limit = this.legendLimit();

    if (!limit || data.length <= limit) return data;

    const visibleCount = Math.max(0, limit - 1);
    const visible = data.slice(0, visibleCount);
    const hiddenSeconds = data.slice(visibleCount).reduce((sum, d) => sum + d.seconds, 0);

    return [...visible, { label: 'Other', seconds: hiddenSeconds }];
  });

  totalSeconds = computed(() => this.displayData().reduce((sum, d) => sum + d.seconds, 0));

  slices = computed<ComputedSlice[]>(() => {
    const total = this.totalSeconds();
    if (total === 0) return [];

    let cumulativeAngle = 0;
    return this.displayData().map((d, i) => {
      const percentage = (d.seconds / total) * 100;
      const angle = (d.seconds / total) * 360;
      const pathD = this.describeArc(cumulativeAngle, cumulativeAngle + angle);
      cumulativeAngle += angle;

      const color =
        d.label === 'Other'
          ? '#6b7280'
          : this.useLanguageColors()
            ? getLanguageColor(d.label, i)
            : GENERIC_COLORS[i % GENERIC_COLORS.length];

      return {
        ...d,
        id: d.label === 'Other' ? 'other' : `${i}-${d.label}`,
        label: this.useLanguageColors() && d.label !== 'Other' ? this.capitalize(d.label) : d.label,
        percentage: Math.round(percentage),
        color,
        pathD,
      };
    });
  });

  private describeArc(startAngle: number, endAngle: number): string {
    const sweep = endAngle - startAngle;

    if (sweep >= 360) {
      const r = 50;
      const cx = 60;
      const cy = 60;

      return `M ${cx - r} ${cy}
            A ${r} ${r} 0 1 0 ${cx + r} ${cy}
            A ${r} ${r} 0 1 0 ${cx - r} ${cy} Z`;
    }

    const r = 50;
    const cx = 60;
    const cy = 60;

    const start = this.polarToCartesian(cx, cy, r, endAngle);
    const end = this.polarToCartesian(cx, cy, r, startAngle);
    const largeArcFlag = sweep > 180 ? '1' : '0';

    return `M ${cx} ${cy} L ${start.x} ${start.y} A ${r} ${r} 0 ${largeArcFlag} 0 ${end.x} ${end.y} Z`;
  }

  private polarToCartesian(cx: number, cy: number, r: number, angleDeg: number) {
    const angleRad = ((angleDeg - 90) * Math.PI) / 180;
    return {
      x: cx + r * Math.cos(angleRad),
      y: cy + r * Math.sin(angleRad),
    };
  }

  private capitalize(str: string): string {
    return str.charAt(0).toUpperCase() + str.slice(1);
  }

  onSliceHover(slice: ComputedSlice) {
    this.hoveredSlice.set(slice);
  }

  onSliceLeave() {
    this.hoveredSlice.set(null);
  }

  formatDuration(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) return `${hours}h ${minutes}m`;
    return `${minutes}m`;
  }

  legendSlices = computed(() => this.slices());
}
