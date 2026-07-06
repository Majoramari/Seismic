import { Component, input, computed } from '@angular/core';

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
export class GoalCard {
  goal = input.required<GoalData>();

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

  timeLeft = computed(() => {
    const g = this.goal();
    if (g.period === 'daily') return 'today';
    return 'this week';
  });

  // SVG circle: circumference for r=18 is ~113. Used to
  // calculate stroke-dashoffset for the progress ring.
  circleOffset = computed(() => {
    const circumference = 113;
    const pct = Math.min(this.goal().percentage, 100);
    return circumference - (circumference * pct) / 100;
  });

  private formatShort(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    if (hours > 0) return `${hours}h`;
    return `${minutes}m`;
  }
}
