import { Component, inject } from '@angular/core';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';
import { NgOptimizedImage } from '@angular/common';
import { LucideAngularModule, Gauge, Trophy, Settings, LogOut, Flame } from 'lucide-angular';
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

  logout() {
    this.auth.logout();
    this.router.navigate(['/login']);
  }
}
