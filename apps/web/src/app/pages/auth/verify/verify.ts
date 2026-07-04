import { Component, inject, signal, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { ApiService } from '../../../core/api/api.service';

interface VerifyResponse {
  newUser: boolean;
  accessToken?: string;
  signupToken?: string;
  email?: string;
  user?: {
    id: string;
    username: string;
    email: string;
  };
}

@Component({
  selector: 'app-verify',
  standalone: true,
  templateUrl: './verify.html',
})
export class Verify implements OnInit {
  private api = inject(ApiService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);

  loading = signal(true);
  error = signal('');

  ngOnInit() {
    const token = this.route.snapshot.queryParamMap.get('token');
    if (!token) {
      this.error.set('Missing token');
      this.loading.set(false);
      return;
    }

    this.api.get<VerifyResponse>('/api/auth/verify', { token }).subscribe({
      next: (data) => {
        if (data.newUser) {
          console.log('New user, signup token:', data.signupToken);
        } else if (data.accessToken) {
          localStorage.setItem('accessToken', data.accessToken);
          this.router.navigate(['/dashboard']);
        }
      },
      error: () => {
        this.error.set('Invalid or expired link');
        this.loading.set(false);
      },
    });
  }
}
