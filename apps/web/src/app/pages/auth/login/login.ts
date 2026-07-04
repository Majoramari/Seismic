import { Component, inject, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../../core/api/api.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './login.html',
})
export class Login {
  private api = inject(ApiService);

  email = signal('');
  loading = signal(false);
  sent = signal(false);
  error = signal('');

  requestMagicLink() {
    this.loading.set(true);
    this.error.set('');

    this.api.post('/api/auth/magic-link', { email: this.email() }).subscribe({
      next: () => {
        this.loading.set(false);
        this.sent.set(true);
      },
      error: (err) => {
        this.loading.set(false);
        this.error.set(err.error?.message || 'Something went wrong');
      },
    });
  }
}
