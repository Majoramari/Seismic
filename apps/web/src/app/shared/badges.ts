export interface BadgeInfo {
  type: string;
  label: string;
  description: string;
  color: string;
}

const BADGES = new Map<string, BadgeInfo>([
  [
    'first_heartbeat',
    {
      type: 'first_heartbeat',
      label: 'First steps',
      description: 'Every legend starts somewhere.',
      color: '#22c55e',
    },
  ],
  [
    'week_streak',
    {
      type: 'week_streak',
      label: 'Week streak',
      description: 'Seven days without missing a beat.',
      color: '#f59e0b',
    },
  ],
  [
    'month_streak',
    {
      type: 'month_streak',
      label: 'Month streak',
      description: 'A month that speaks for itself.',
      color: '#ef4444',
    },
  ],
  [
    'night_owl',
    {
      type: 'night_owl',
      label: 'Night owl',
      description: 'The stars witnessed every commit.',
      color: '#6366f1',
    },
  ],
  [
    'early_bird',
    {
      type: 'early_bird',
      label: 'Early bird',
      description: 'Ahead of sunrise.',
      color: '#eab308',
    },
  ],
  [
    'polyglot',
    {
      type: 'polyglot',
      label: 'Polyglot',
      description: 'No single language tells the whole story.',
      color: '#06b6d4',
    },
  ],
  [
    'century',
    {
      type: 'century',
      label: 'Century',
      description: '100 hours of coding. Time well invested.',
      color: '#d97757',
    },
  ],
  [
    'supporter',
    {
      type: 'supporter',
      label: 'Supporter',
      description: 'Keeping Seismic alive, one donation at a time.',
      color: '#ec4899',
    },
  ],
  [
    'contributor',
    {
      type: 'contributor',
      label: 'Contributor',
      description: 'Left a mark on Seismic.',
      color: '#8b5cf6',
    },
  ],
  [
    'maintainer',
    {
      type: 'maintainer',
      label: 'Maintainer',
      description: 'Maintaining what others help build.',
      color: '#fbbf24',
    },
  ],
]);

export function badgeInfo(type: string): BadgeInfo {
  return (
    BADGES.get(type) ?? {
      type,
      label: type,
      description: '',
      color: '#6b7280',
    }
  );
}
