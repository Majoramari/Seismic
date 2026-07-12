import { Component, OnDestroy, input, computed, signal } from '@angular/core';

export interface GoalData {
  id: string;
  scope: string;
  scopeValue: string | null;
  period: string;
  targetSeconds: number;
  progressSeconds: number;
  percentage: number;
}

@Component({
  selector: 'app-goal-card',
  standalone: true,
  templateUrl: './goal-card.html',
})
export class GoalCard implements OnDestroy {
  goal = input.required<GoalData>();

  private now = signal(new Date());
  private resetClock = setInterval(() => this.now.set(new Date()), 60_000);

  isComplete = computed(() => this.goal().percentage >= 100);

  label = computed(() => {
    const g = this.goal();
    if (g.scope === 'language') return `Language: ${g.scopeValue}`;
    if (g.scope === 'project') return `Project: ${g.scopeValue}`;
    return 'All programming activity';
  });

  progressText = computed(() => {
    const g = this.goal();
    return `${this.formatShort(g.progressSeconds)}/${this.formatShort(g.targetSeconds)}`;
  });

  periodLabel = computed(() => {
    const g = this.goal();
    if (g.period === 'daily') return 'Daily goal';
    if (g.period === 'weekly') return 'Weekly goal';
    return 'Monthly goal';
  });

  remainingText = computed(() => {
    const g = this.goal();
    const now = this.now();
    const resetAt = this.nextResetAt(g.period, now);
    return `${this.formatResetDuration(resetAt.getTime() - now.getTime())} left ${this.periodRemainingLabel(g.period)}`;
  });

  barPercentage = computed(() => Math.min(this.goal().percentage, 100));

  circleOffset = computed(() => {
    const circumference = 113;
    const pct = Math.min(this.goal().percentage, 100);
    return circumference - (circumference * pct) / 100;
  });

  ngOnDestroy(): void {
    clearInterval(this.resetClock);
  }

  private formatShort(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) return `${hours}h${minutes > 0 ? ' ' + minutes + 'm' : ''}`;
    return `${minutes}m`;
  }

  private periodRemainingLabel(period: string): string {
    if (period === 'daily') return 'today';
    if (period === 'weekly') return 'this week';
    return 'this month';
  }

  private nextResetAt(period: string, now: Date): Date {
    if (period === 'weekly') {
      const day = now.getDay();
      const daysUntilMonday = day === 0 ? 1 : 8 - day;
      return new Date(now.getFullYear(), now.getMonth(), now.getDate() + daysUntilMonday);
    }

    if (period === 'monthly') {
      return new Date(now.getFullYear(), now.getMonth() + 1, 1);
    }

    return new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1);
  }

  private formatResetDuration(milliseconds: number): string {
    const totalMinutes = Math.max(Math.ceil(milliseconds / 60_000), 0);
    const days = Math.floor(totalMinutes / 1_440);
    const hours = Math.floor((totalMinutes % 1_440) / 60);
    const minutes = totalMinutes % 60;

    if (days > 0) return `${days}d${hours > 0 ? ' ' + hours + 'h' : ''}`;
    if (hours > 0) return `${hours}h${minutes > 0 ? ' ' + minutes + 'm' : ''}`;
    return `${minutes}m`;
  }
}
