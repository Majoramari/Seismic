import { Component, inject, signal, OnInit } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { ApiService } from '../../../core/api/api.service';
import { AuthService } from '../../../core/auth/auth.service';
import { debounceTime, Subject } from 'rxjs';

interface VerifyResponse {
  newUser: boolean;
  accessToken?: string;
  signupToken?: string;
  email?: string;
}

interface SignupResponse {
  accessToken: string;
  user: { id: string; username: string; email: string };
}

interface CheckUsernameResponse {
  available: boolean;
  reason?: string;
}

type OnboardingStep = 'username' | 'profile';
type UsernameStatus = 'idle' | 'checking' | 'available' | 'taken' | 'invalid';

@Component({
  selector: 'app-verify',
  standalone: true,
  imports: [FormsModule],
  templateUrl: './verify.html',
})
export class Verify implements OnInit {
  private api = inject(ApiService);
  private auth = inject(AuthService);
  private route = inject(ActivatedRoute);
  private router = inject(Router);

  loading = signal(true);
  error = signal('');
  isNewUser = signal(false);
  step = signal<OnboardingStep>('username');

  signupToken = '';
  email = signal('');

  username = signal('');
  usernameStatus = signal<UsernameStatus>('idle');
  displayName = signal('');
  signupLoading = signal(false);
  signupError = signal('');

  private usernameCheck$ = new Subject<string>();

  ngOnInit() {
    this.usernameCheck$.pipe(debounceTime(400)).subscribe((value) => {
      this.checkUsername(value);
    });

    const token = this.route.snapshot.queryParamMap.get('token');
    if (!token) {
      this.error.set('Missing token');
      this.loading.set(false);
      return;
    }

    this.api.get<VerifyResponse>('/api/auth/verify', { token }).subscribe({
      next: (data) => {
        this.loading.set(false);

        if (data.newUser && data.signupToken) {
          this.isNewUser.set(true);
          this.signupToken = data.signupToken;
          this.email.set(data.email ?? '');
        } else if (data.accessToken) {
          this.auth.setToken(data.accessToken);
          this.router.navigate(['/dashboard']);
        }
      },
      error: () => {
        this.error.set('Invalid or expired link');
        this.loading.set(false);
      },
    });
  }

  onUsernameInput(value: string) {
    this.username.set(value);
    if (value.length < 3) {
      this.usernameStatus.set('idle');
      return;
    }
    this.usernameStatus.set('checking');
    this.usernameCheck$.next(value);
  }

  private checkUsername(value: string) {
    this.api.get<CheckUsernameResponse>('/api/auth/check-username', { username: value }).subscribe({
      next: (data) => {
        if (!data.available && data.reason === 'invalid_format') {
          this.usernameStatus.set('invalid');
        } else {
          this.usernameStatus.set(data.available ? 'available' : 'taken');
        }
      },
      error: () => this.usernameStatus.set('idle'),
    });
  }

  goToProfileStep() {
    if (this.usernameStatus() !== 'available') return;
    this.step.set('profile');
  }

  backToUsernameStep() {
    this.step.set('username');
  }

  completeSignup() {
    this.signupLoading.set(true);
    this.signupError.set('');

    this.api
      .post<SignupResponse>('/api/auth/complete-signup', {
        signupToken: this.signupToken,
        username: this.username(),
        displayName: this.displayName(),
      })
      .subscribe({
        next: (data) => {
          this.auth.setToken(data.accessToken);
          this.router.navigate(['/dashboard']);
        },
        error: (err) => {
          this.signupLoading.set(false);
          this.signupError.set(err.error?.message || 'Something went wrong');
        },
      });
  }
}
