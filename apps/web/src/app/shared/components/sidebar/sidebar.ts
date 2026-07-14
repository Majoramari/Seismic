import { Component, inject } from '@angular/core';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';
import { NgOptimizedImage } from '@angular/common';
import {
  LucideAngularModule,
  Gauge,
  Trophy,
  Settings,
  LogOut,
  Flame,
  BookOpen,
  FolderGit2,
} from 'lucide-angular';
import { AuthService } from '../../../core/auth/auth.service';
import { UserService } from '../../../core/auth/user.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [RouterLink, RouterLinkActive, LucideAngularModule, NgOptimizedImage],
  templateUrl: './sidebar.html',
})
export class Sidebar {
  private auth = inject(AuthService);
  private router = inject(Router);
  userService = inject(UserService);

  readonly GaugeIcon = Gauge;
  readonly TrophyIcon = Trophy;
  readonly SettingsIcon = Settings;
  readonly LogOutIcon = LogOut;
  readonly FlameIcon = Flame;
  readonly DocsIcon = BookOpen;
  readonly ProjectsIcon = FolderGit2;

  streakTier(streak: number): string {
    if (streak >= 30) return 'streak-tier-legendary';
    if (streak >= 14) return 'streak-tier-hot';
    if (streak >= 7) return 'streak-tier-warm';
    return 'streak-tier-new';
  }

  logout() {
    this.auth.logout();
    void this.router.navigate(['/login']);
  }
}
