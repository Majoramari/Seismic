import { Injectable, inject, signal } from '@angular/core';
import { ApiService } from '../api/api.service';
import { retry, timer } from 'rxjs';

interface CurrentUser {
  username: string;
  displayName: string;
  email: string;
  bio: string;
  website: string;
  profileImage: string | null;
  country: string | null;
}

interface StatsSummary {
  currentStreak: number;
}

@Injectable({ providedIn: 'root' })
export class UserService {
  private api = inject(ApiService);

  user = signal<CurrentUser | null>(null);
  streak = signal<number>(0);

  load() {
    this.api
      .get<CurrentUser>('/api/auth/me')
      .pipe(retry({ count: 2, delay: (_, retryIndex) => timer(retryIndex * 500) }))
      .subscribe({
        next: (data) => this.user.set(data),
        error: (err) => console.error('Failed to load user after retries:', err),
      });

    this.api
      .get<StatsSummary>('/api/stats/summary', { range: 'today' })
      .pipe(retry({ count: 2, delay: (_, retryIndex) => timer(retryIndex * 500) }))
      .subscribe({
        next: (data) => this.streak.set(data.currentStreak),
        error: (err) => console.error('Failed to load streak after retries:', err),
      });
  }
}
