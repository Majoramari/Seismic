import { Component, inject } from '@angular/core';
import { Router, RouterLink } from '@angular/router';
import { AuthService } from '../../../core/auth/auth.service';
import { NgOptimizedImage } from '@angular/common';

@Component({
  selector: 'app-navbar',
  standalone: true,
  imports: [RouterLink, NgOptimizedImage],
  templateUrl: './navbar.html',
})
export class Navbar {
  auth = inject(AuthService);
  private router = inject(Router);

  logout() {
    this.auth.logout();
    void this.router.navigate(['/login']);
  }
}
