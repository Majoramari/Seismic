import { Component, inject, signal, OnInit } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../core/api/api.service';
import { RouterLink } from '@angular/router';
import { LucideAngularModule, Flame } from 'lucide-angular';

interface LeaderboardEntry {
  rank: number;
  username: string;
  seconds: number;
  topLanguage: string;
  isYou: boolean;
  profileImage: string | null;
  streak: number;
}

interface LeaderboardResult {
  entries: LeaderboardEntry[];
  yourRank: number | null;
}

type RangeOption = 'today' | 'week' | 'month' | 'all';

@Component({
  selector: 'app-leaderboard',
  standalone: true,
  imports: [FormsModule, RouterLink, LucideAngularModule],
  templateUrl: './leaderboard.html',
})
export class Leaderboard implements OnInit {
  private api = inject(ApiService);

  readonly FlameIcon = Flame;

  entries = signal<LeaderboardEntry[]>([]);
  yourRank = signal<number | null>(null);
  loading = signal(true);
  range = signal<RangeOption>('today');
  search = signal('');

  filteredEntries = signal<LeaderboardEntry[]>([]);

  // Cache results per range for 60 seconds, so switching
  // tabs repeatedly doesn't refetch data we already have.
  private cache = new Map<RangeOption, { data: LeaderboardResult; fetchedAt: number }>();
  private readonly CACHE_TTL = 60_000;

  ngOnInit() {
    this.loadLeaderboard();
  }

  setRange(range: RangeOption) {
    this.range.set(range);
    this.loadLeaderboard();
  }

  onSearchChange(value: string) {
    this.search.set(value);
    const query = value.toLowerCase();
    this.filteredEntries.set(
      this.entries().filter((e) => e.username.toLowerCase().includes(query)),
    );
  }

  formatSeconds(seconds: number): string {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const secs = seconds % 60;
    if (hours > 0) return `${hours}h ${minutes}m ${secs}s`;
    if (minutes > 0) return `${minutes}m ${secs}s`;
    return `${secs}s`;
  }

  private loadLeaderboard() {
    const currentRange = this.range();
    const cached = this.cache.get(currentRange);

    if (cached && Date.now() - cached.fetchedAt < this.CACHE_TTL) {
      this.applyResult(cached.data);
      return;
    }

    this.loading.set(true);
    this.api.get<LeaderboardResult>('/api/leaderboard', { range: currentRange }).subscribe({
      next: (data) => {
        this.cache.set(currentRange, { data, fetchedAt: Date.now() });
        this.applyResult(data);
        this.loading.set(false);
      },
      error: () => this.loading.set(false),
    });
  }

  private applyResult(data: LeaderboardResult) {
    this.entries.set(data.entries ?? []);
    this.filteredEntries.set(data.entries ?? []);
    this.yourRank.set(data.yourRank);
  }

  streakTier(streak: number): string {
    if (streak >= 30) return 'streak-tier-legendary';
    if (streak >= 14) return 'streak-tier-hot';
    if (streak >= 7) return 'streak-tier-warm';
    return 'streak-tier-new';
  }
}
