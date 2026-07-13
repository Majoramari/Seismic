import { Component, OnInit, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../core/api/api.service';
import { ToastService } from '../../core/toast/toast.service';
import { UserService } from '../../core/auth/user.service';
import { Heatmap } from '../../shared/components/heatmap/heatmap';

interface HeatmapDay {
  date: string;
  seconds: number;
}

interface ProfileStats {
  totalSeconds: number | null;
  topProject: string | null;
  topLanguage: string | null;
  topOS: string | null;
  topEditor: string | null;
  currentStreak: number | null;
}

interface ProfileVisibility {
  hideTime: boolean;
  hideProjects: boolean;
  hideLanguages: boolean;
}

interface ProfileData {
  username: string;
  displayName: string;
  accountEmail?: string;
  contactEmail: string;
  bio: string;
  website: string;
  profileImage: string | null;
  country: string | null;
  createdAt?: string;
  isOwner: boolean;
  visibility: ProfileVisibility;
  stats: ProfileStats;
  heatmap: HeatmapDay[];
}

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [FormsModule, Heatmap],
  templateUrl: './profile.html',
  styleUrl: './profile.css',
})
export class Profile implements OnInit {
  private api = inject(ApiService);
  private toast = inject(ToastService);
  private userService = inject(UserService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);

  loading = signal(true);
  saving = signal(false);
  notFound = signal(false);
  profile = signal<ProfileData | null>(null);
  profileImage = signal<string | null>(null);

  username = '';
  displayName = '';
  accountEmail = '';
  contactEmail = '';
  bio = '';
  website = '';

  ngOnInit(): void {
    this.route.paramMap.subscribe((params) => {
      const username = params.get('username');

      if (username) {
        this.loadProfile(username);
      }
    });
  }

  loadProfile(username: string): void {
    this.loading.set(true);
    this.notFound.set(false);

    this.api.get<ProfileData>(`/api/users/${encodeURIComponent(username)}`).subscribe({
      next: (profile) => {
        this.profile.set(profile);
        this.username = profile.username ?? '';
        this.displayName = profile.displayName ?? '';
        this.accountEmail = profile.accountEmail ?? '';
        this.contactEmail = profile.contactEmail ?? '';
        this.bio = profile.bio ?? '';
        this.website = profile.website ?? '';
        this.profileImage.set(profile.profileImage ?? null);
        this.loading.set(false);
      },
      error: (error) => {
        this.loading.set(false);
        this.notFound.set(true);

        if (error.status === 403) {
          this.toast.error('This profile is private');
        }
      },
    });
  }

  onImageSelected(event: Event): void {
    if (!this.profile()?.isOwner) return;

    const input = event.target as HTMLInputElement;
    const file = input.files?.[0];

    if (!file) return;

    if (!file.type.startsWith('image/')) {
      this.toast.error('Please select an image file');
      input.value = '';
      return;
    }

    if (file.size > 2 * 1024 * 1024) {
      this.toast.error('Profile images must be smaller than 2 MB');
      input.value = '';
      return;
    }

    const reader = new FileReader();

    reader.onload = () => {
      this.profileImage.set(reader.result as string);
    };

    reader.onerror = () => {
      this.toast.error('Failed to read the selected image');
    };

    reader.readAsDataURL(file);
  }

  removeImage(): void {
    if (!this.profile()?.isOwner) return;
    this.profileImage.set(null);
  }

  saveProfile(): void {
    if (!this.profile()?.isOwner) return;

    const previousUsername = this.profile()!.username;
    const username = this.username.trim();
    const displayName = this.displayName.trim();
    const website = this.website.trim();

    if (username.length < 3) {
      this.toast.error('Username must contain at least 3 characters');
      return;
    }

    if (!/^[a-zA-Z0-9_-]+$/.test(username)) {
      this.toast.error('Username can only contain letters, numbers, underscores, and hyphens');
      return;
    }

    if (website && !this.isValidWebsite(website)) {
      this.toast.error('Please enter a valid website address');
      return;
    }

    this.saving.set(true);

    this.api
      .put<ProfileData>('/api/auth/profile', {
        username,
        displayName,
        bio: this.bio.trim(),
        website,
        contactEmail: this.contactEmail.trim(),
        profileImage: this.profileImage(),
      })
      .subscribe({
        next: (profile) => {
          this.profile.set(profile);
          this.username = profile.username;
          this.displayName = profile.displayName ?? '';
          this.bio = profile.bio ?? '';
          this.website = profile.website ?? '';
          this.profileImage.set(profile.profileImage ?? null);
          this.saving.set(false);
          this.userService.load();
          this.toast.success('Profile updated');

          if (previousUsername !== profile.username) {
            void this.router.navigate(['/p', profile.username], {
              replaceUrl: true,
            });
          }
        },
        error: (error) => {
          this.saving.set(false);
          this.toast.error(error.error?.message ?? 'Failed to update your profile');
        },
      });
  }

  formatSeconds(seconds: number | null): string {
    if (seconds === null) return 'Hidden';
    if (seconds < 60) return '< 1m';

    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    if (hours > 0 && minutes > 0) return `${hours}h ${minutes}m`;
    if (hours > 0) return `${hours}h`;

    return `${minutes}m`;
  }

  profileInitial(): string {
    const value = this.displayName.trim() || this.username.trim();

    return value.charAt(0).toUpperCase() || '?';
  }

  joinedLabel(): string {
    const createdAt = this.profile()?.createdAt;
    if (!createdAt) return '';

    return new Date(createdAt).toLocaleDateString('en-US', {
      month: 'long',
      year: 'numeric',
    });
  }

  normalizedWebsite(): string {
    const website = this.profile()?.website?.trim();
    if (!website) return '';

    return website.startsWith('http://') || website.startsWith('https://')
      ? website
      : `https://${website}`;
  }

  private isValidWebsite(value: string): boolean {
    try {
      const url = new URL(
        value.startsWith('http://') || value.startsWith('https://') ? value : `https://${value}`,
      );

      return Boolean(url.hostname);
    } catch {
      return false;
    }
  }

  private readonly labelOverrides = new Map<string, string>([
    ['macos', 'macOS'],
    ['ios', 'iOS'],
    ['linux', 'Linux'],
    ['windows', 'Windows'],
  ]);

  formatDisplayLabel(value: string | null | undefined): string {
    if (!value) return '—';
    const trimmed = value.trim();
    if (!trimmed) return '—';

    const override = this.labelOverrides.get(trimmed.toLowerCase());
    if (override) return override;

    return trimmed.charAt(0).toUpperCase() + trimmed.slice(1);
  }
}
