import { Component, inject, signal } from '@angular/core';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';
import { NgOptimizedImage } from '@angular/common';
import {
  LucideAngularModule,
  Gauge,
  Trophy,
  Settings,
  LogOut,
  Flame,
  BarChart3,
  FolderKanban,
  User,
  ChevronDown,
  Monitor,
  Moon,
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
  readonly StatsIcon = BarChart3;
  readonly ProjectsIcon = FolderKanban;
  readonly ProfileIcon = User;
  readonly ChevronDownIcon = ChevronDown;
  readonly MonitorIcon = Monitor;
  readonly MoonIcon = Moon;

  darkMode = signal(true);

  toggleDarkMode() {
    this.darkMode.update((v) => !v);
  }

  logout() {
    this.auth.logout();
    void this.router.navigate(['/login']);
  }
}

