import { inject, Injectable } from '@angular/core';
import { Observable } from 'rxjs';
import { ApiService } from '../../core/api/api.service';

// ── Response interfaces (mirror models.ProfileResponse from the Go backend) ──

export interface HeatmapCell {
  date: string;
  level: number; // 0-4
}

export interface ProfileInfoField {
  key: string;
  label: string;
  completed: boolean;
}

export interface Achievement {
  key: string;
  title: string;
  description: string;
  badgeClass: string;
  earnedAt: string; // "Jan 02, 2006"
}

export interface ActivityLogItem {
  kind: string;
  text: string;
  at: string; // RFC3339
}

export interface ProfileResponse {
  // Identity
  username: string;
  email: string;
  firstName: string;
  avatarUrl: string;

  // Bio
  role: string;
  location: string;
  university: string;
  bio: string;
  timeZone: string;
  website: string;
  gender: string;
  languages: string[];

  // Date strings
  joinDate: string;
  lastActive: string;
  memberFor: string;

  // Coding metrics
  totalCodingSeconds: number;
  totalActiveDays: number;
  currentStreak: number;
  maxStreak: number;

  // Problem solving (placeholder)
  solved: number;
  totalProblems: number;
  attempting: number;

  // Personal info completion
  completionPercent: number;
  infoFields: ProfileInfoField[];

  // Rich data
  heatmap: HeatmapCell[][];
  recentActivity: ActivityLogItem[];
  achievements: Achievement[];
}

@Injectable({ providedIn: 'root' })
export class ProfileService {
  private api = inject(ApiService);

  /** Fetches the full profile for the currently logged-in user. */
  getProfile(): Observable<ProfileResponse> {
    return this.api.get<ProfileResponse>('/api/profile');
  }

  /** Updates the allowed profile fields. */
  updateProfile(data: UpdateProfileRequest): Observable<any> {
    return this.api.patch('/api/profile', data);
  }
}

export interface UpdateProfileRequest {
  firstName?: string | null;
  bio?: string | null;
  location?: string | null;
  role?: string | null;
  university?: string | null;
  website?: string | null;
  gender?: string | null;
  languages?: string[] | null;
  timeZone?: string | null;
}
