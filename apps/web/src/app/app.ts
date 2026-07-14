import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { NavigationEnd, Router, RouterOutlet } from '@angular/router';
import { Navbar } from './shared/components/navbar/navbar';
import { Sidebar } from './shared/components/sidebar/sidebar';
import { ToastContainer } from './shared/components/toast/toast';
import { AuthService } from './core/auth/auth.service';
import { UserService } from './core/auth/user.service';
import { filter } from 'rxjs';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [RouterOutlet, Navbar, Sidebar, ToastContainer],
  templateUrl: './app.html',
})
export class App implements OnInit {
  auth = inject(AuthService);
  private userService = inject(UserService);
  private router = inject(Router);

  private currentUrl = signal(this.router.url);
  showAuthenticatedShell = computed(() => {
    return this.auth.isLoggedIn() && !this.isOnboardingRoute(this.currentUrl());
  });

  showNavbar = computed(() => !this.isOnboardingRoute(this.currentUrl()));

  ngOnInit() {
    if (this.auth.isLoggedIn()) {
      this.userService.load();
    }

    this.router.events
      .pipe(filter((event): event is NavigationEnd => event instanceof NavigationEnd))
      .subscribe((event) => this.currentUrl.set(event.urlAfterRedirects));
  }

  private isOnboardingRoute(url: string): boolean {
    return url.split('?')[0] === '/verify';
  }
}
