import { Component, inject } from '@angular/core';
import { Router, RouterLink, RouterLinkActive } from '@angular/router';
import { LucideAngularModule, Gauge, Trophy, Settings, LogOut } from 'lucide-angular';
import { AuthService } from '../../../core/auth/auth.service';

@Component({
  selector: 'app-sidebar',
  standalone: true,
  imports: [RouterLink, RouterLinkActive, LucideAngularModule],
  templateUrl: './sidebar.html',
})
export class Sidebar {
  private auth = inject(AuthService);
  private router = inject(Router);

  readonly GaugeIcon = Gauge;
  readonly TrophyIcon = Trophy;
  readonly SettingsIcon = Settings;
  readonly LogOutIcon = LogOut;

  logout() {
    this.auth.logout();
    this.router.navigate(['/login']);
  }
}
