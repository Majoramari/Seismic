import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { RouterLink, Router } from '@angular/router';
import { NgOptimizedImage } from '@angular/common';
import { LucideAngularModule, Mailbox } from 'lucide-angular';
import { ApiService } from '../../../core/api/api.service';
import { ToastService } from '../../../core/toast/toast.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule, RouterLink, NgOptimizedImage, LucideAngularModule],
  templateUrl: './login.html',
})
export class Login {
  private api = inject(ApiService);
  private toast = inject(ToastService);
  private router = inject(Router);

  readonly MailboxIcon = Mailbox;

  email = signal('');
  loading = signal(false);
  sent = signal(false);

  requestMagicLink() {
    if (!this.email() || this.loading()) return;

    this.loading.set(true);

    this.api.post('/api/auth/magic-link', { email: this.email() }).subscribe({
      next: (res: any) => {
        if (res.data?.devToken) {
          void this.router.navigate(['/verify'], { queryParams: { token: res.data.devToken } });
          return;
        }
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
