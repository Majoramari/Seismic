import { Component, ElementRef, computed, effect, input, signal, viewChild } from '@angular/core';

interface HeatmapDay {
  date: string;
  seconds: number;
}

interface DisplayDay {
  date: string;
  seconds: number;
  level: number;
  isOutsidePeriod: boolean;
}

interface MonthLabel {
  name: string;
  weekIndex: number;
}

@Component({
  selector: 'app-heatmap',
  standalone: true,
  templateUrl: './heatmap.html',
})
export class Heatmap {
  data = input.required<HeatmapDay[]>();

  private selectedYear = signal(new Date().getFullYear());
  private heatmapScroll = viewChild<ElementRef<HTMLDivElement>>('heatmapScroll');

  private activityByDate = computed(() => {
    const activity = new Map<string, number>();

    for (const item of this.data()) {
      const date = item.date.slice(0, 10);
      activity.set(date, (activity.get(date) ?? 0) + item.seconds);
    }

    return activity;
  });

  availableYears = computed(() => {
    const currentYear = new Date().getFullYear();
    let firstYear = currentYear;
    let lastYear = currentYear;

    for (const dateString of this.activityByDate().keys()) {
      const year = this.parseDate(dateString).getFullYear();
      firstYear = Math.min(firstYear, year);
      lastYear = Math.max(lastYear, year);
    }

    const years: number[] = [];

    for (let year = lastYear; year >= firstYear; year--) {
      years.push(year);
    }

    return years;
  });

  activeYear = computed(() => {
    const selected = this.selectedYear();
    const years = this.availableYears();

    return years.includes(selected) ? selected : years[0];
  });

  activePeriod = computed(() => {
    const year = this.activeYear();
    const currentYear = new Date().getFullYear();

    if (year === currentYear) {
      let end = this.startOfDay(new Date());

      for (const dateString of this.activityByDate().keys()) {
        const activityDate = this.parseDate(dateString);

        if (activityDate.getFullYear() === currentYear && activityDate > end) {
          end = activityDate;
        }
      }

      const start = new Date(end);
      start.setDate(start.getDate() - 364);

      return {
        start,
        end,
        label: 'in the last year',
      };
    }

    return {
      start: new Date(year, 0, 1),
      end: new Date(year, 11, 31),
      label: `in ${year}`,
    };
  });

  yearTotal = computed(() => {
    const period = this.activePeriod();
    let total = 0;

    for (const [dateString, seconds] of this.activityByDate()) {
      const date = this.parseDate(dateString);

      if (this.isInPeriod(date, period.start, period.end)) {
        total += seconds;
      }
    }

    return total;
  });

  activeDays = computed(() => {
    const period = this.activePeriod();
    let count = 0;

    for (const [dateString, seconds] of this.activityByDate()) {
      const date = this.parseDate(dateString);

      if (seconds > 0 && this.isInPeriod(date, period.start, period.end)) {
        count++;
      }
    }

    return count;
  });

  monthLabels = computed(() => {
    const period = this.activePeriod();
    const gridStart = this.getGridStart(period.start);
    const labels: MonthLabel[] = [];
    const cursor = new Date(period.start.getFullYear(), period.start.getMonth(), 1);

    if (cursor < period.start) {
      cursor.setMonth(cursor.getMonth() + 1);
    }

    while (cursor <= period.end) {
      labels.push({
        name: cursor.toLocaleDateString('en-US', {
          month: 'short',
        }),
        weekIndex: Math.floor(this.daysBetween(gridStart, cursor) / 7),
      });

      cursor.setMonth(cursor.getMonth() + 1);
    }

    return labels;
  });

  weeks = computed(() => {
    const activity = this.activityByDate();
    const period = this.activePeriod();
    const gridStart = this.getGridStart(period.start);
    const gridEnd = this.getGridEnd(period.end);
    const days: DisplayDay[] = [];
    const cursor = new Date(gridStart);

    while (cursor <= gridEnd) {
      const date = this.formatDateKey(cursor);
      const seconds = activity.get(date) ?? 0;
      const isBeforePeriod = cursor < period.start;
      const isAfterPeriod = cursor > period.end;

      // Padded days stay hidden unless they contain recorded activity.
      const isOutsidePeriod = (isBeforePeriod || isAfterPeriod) && seconds <= 0;

      days.push({
        date,
        seconds,
        level: this.getLevel(seconds),
        isOutsidePeriod,
      });

      cursor.setDate(cursor.getDate() + 1);
    }

    const weeks: DisplayDay[][] = [];

    for (let index = 0; index < days.length; index += 7) {
      weeks.push(days.slice(index, index + 7));
    }

    return weeks;
  });

  weekCount = computed(() => this.weeks().length);

  totalLabel = computed(
    () => `${this.formatDuration(this.yearTotal())} coded ${this.activePeriod().label}`,
  );

  constructor() {
    effect(() => {
      this.weeks();
      const element = this.heatmapScroll()?.nativeElement;

      if (!element) return;

      window.setTimeout(() => {
        element.scrollLeft = element.scrollWidth - element.clientWidth;
      });
    });
  }

  setYear(year: number): void {
    this.selectedYear.set(year);
  }

  formatTooltip(day: DisplayDay): string {
    if (day.isOutsidePeriod) return '';

    const hours = Math.floor(day.seconds / 3600);
    const minutes = Math.floor((day.seconds % 3600) / 60);
    const seconds = day.seconds % 60;
    const parts: string[] = [];

    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    if (seconds > 0) parts.push(`${seconds}s`);

    const time = parts.length === 0 ? 'No activity' : parts.join(' ');

    const date = this.parseDate(day.date).toLocaleDateString('en-US', {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });

    return `${time} on ${date}`;
  }

  formatDuration(seconds: number): string {
    if (seconds <= 0) return '0m';

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (hours > 0) {
      return `${hours}h${minutes > 0 ? ` ${minutes}m` : ''}`;
    }

    return `${minutes}m`;
  }

  private getLevel(seconds: number): number {
    if (seconds <= 0) return 0;
    if (seconds < 1800) return 1;
    if (seconds < 3600) return 2;
    if (seconds < 7200) return 3;

    return 4;
  }

  private getGridStart(date: Date): Date {
    const start = new Date(date);
    start.setDate(start.getDate() - start.getDay());

    return start;
  }

  private getGridEnd(date: Date): Date {
    const end = new Date(date);
    end.setDate(end.getDate() + (6 - end.getDay()));

    return end;
  }

  private startOfDay(date: Date): Date {
    return new Date(date.getFullYear(), date.getMonth(), date.getDate());
  }

  private isInPeriod(date: Date, start: Date, end: Date): boolean {
    return date >= start && date <= end;
  }

  private parseDate(date: string): Date {
    const [year, month, day] = date.slice(0, 10).split('-').map(Number);

    return new Date(year, month - 1, day);
  }

  private formatDateKey(date: Date): string {
    const year = date.getFullYear();
    const month = `${date.getMonth() + 1}`.padStart(2, '0');
    const day = `${date.getDate()}`.padStart(2, '0');

    return `${year}-${month}-${day}`;
  }

  private daysBetween(start: Date, end: Date): number {
    const startUtc = Date.UTC(start.getFullYear(), start.getMonth(), start.getDate());

    const endUtc = Date.UTC(end.getFullYear(), end.getMonth(), end.getDate());

    return Math.round((endUtc - startUtc) / 86_400_000);
  }
}
