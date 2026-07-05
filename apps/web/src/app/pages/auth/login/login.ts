import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink } from '@angular/router';
import { NgOptimizedImage } from '@angular/common';
import { ApiService } from '../../../core/api/api.service';
import { ToastService } from '../../../core/toast/toast.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule, RouterLink, NgOptimizedImage],
  templateUrl: './login.html',
})
export class Login {
  private api = inject(ApiService);
  private toast = inject(ToastService);

  email = signal('');
  loading = signal(false);
  sent = signal(false);

  requestMagicLink() {
    if (!this.email() || this.loading()) return;

    this.loading.set(true);

    this.api.post('/api/auth/magic-link', { email: this.email() }).subscribe({
      next: () => {
        this.loading.set(false);
        this.sent.set(true);
      },
      error: (err) => {
        this.loading.set(false);
        this.toast.error(err.error?.message || 'Something went wrong. Check your connection.');
      },
    });
  }
}
