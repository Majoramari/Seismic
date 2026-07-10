import { Component, input, computed } from '@angular/core';
import { getLanguageColor } from './language-colors';

export interface PieSlice {
  label: string;
  seconds: number;
}

interface ComputedSlice extends PieSlice {
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

  totalSeconds = computed(() => this.data().reduce((sum, d) => sum + d.seconds, 0));

  slices = computed<ComputedSlice[]>(() => {
    const total = this.totalSeconds();
    if (total === 0) return [];

    let cumulativeAngle = 0;
    return this.data().map((d, i) => {
      const percentage = (d.seconds / total) * 100;
      const angle = (d.seconds / total) * 360;
      const pathD = this.describeArc(cumulativeAngle, cumulativeAngle + angle);
      cumulativeAngle += angle;

      const color = this.useLanguageColors()
        ? getLanguageColor(d.label, i)
        : GENERIC_COLORS[i % GENERIC_COLORS.length];

      return {
        ...d,
        label: this.useLanguageColors() ? this.capitalize(d.label) : d.label,
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
}
