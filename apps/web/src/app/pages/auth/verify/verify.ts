import { Component, inject, signal, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { NgOptimizedImage } from '@angular/common';
import { ApiService } from '../../../core/api/api.service';
import { AuthService } from '../../../core/auth/auth.service';
import { UserService } from '../../../core/auth/user.service';
import { debounceTime, Subject, Subscription } from 'rxjs';

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

interface ImportProgress {
  status: string;
  imported: number;
  total: number;
  error?: string;
}

type OnboardingStep = 'username' | 'profile' | 'import';
type ImportProvider = 'wakatime' | 'hackatime';
type UsernameStatus = 'idle' | 'checking' | 'available' | 'taken' | 'invalid';

@Component({
  selector: 'app-verify',
  standalone: true,
  imports: [FormsModule, NgOptimizedImage],
  templateUrl: './verify.html',
})
export class Verify implements OnInit, OnDestroy {
  private api = inject(ApiService);
  private auth = inject(AuthService);
  private userService = inject(UserService);
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
  accountCreated = signal(false);

  selectedImportProvider = signal<ImportProvider | null>(null);
  wakaTimeApiKey = '';
  selectedFile: File | null = null;
  importing = signal(false);
  importStarted = signal(false);
  importError = signal('');
  importProgress = signal<ImportProgress | null>(null);

  private usernameCheck$ = new Subject<string>();
  private usernameCheckSub?: Subscription;
  private importPollInterval: ReturnType<typeof setInterval> | null = null;

  ngOnInit() {
    this.usernameCheckSub = this.usernameCheck$.pipe(debounceTime(400)).subscribe((value) => {
      this.checkUsername(value);
    });

    const token = this.route.snapshot.queryParamMap.get('token');
    if (!token) {
      if (this.auth.isLoggedIn()) {
        void this.router.navigate(['/dashboard']);
        return;
      }

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
          this.userService.load();
          this.router.navigate(['/dashboard']);
        }
      },
      error: () => {
        if (this.auth.isLoggedIn()) {
          void this.router.navigate(['/dashboard']);
          return;
        }

        this.error.set('Invalid or expired link');
        this.loading.set(false);
      },
    });
  }

  ngOnDestroy() {
    this.usernameCheckSub?.unsubscribe();
    this.stopPolling();
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
    if (this.signupLoading() || this.accountCreated()) return;

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
          this.userService.load();
          this.accountCreated.set(true);
          this.signupLoading.set(false);
          this.step.set('import');
        },
        error: (err) => {
          this.signupLoading.set(false);
          this.signupError.set(err.error?.message || 'Something went wrong');
        },
      });
  }

  selectImportProvider(provider: ImportProvider) {
    if (this.importing()) return;

    this.selectedImportProvider.set(provider);
    this.selectedFile = null;
    this.importStarted.set(false);
    this.importError.set('');
    this.importProgress.set(null);
    this.stopPolling();
  }

  onImportFileSelected(event: Event) {
    const input = event.target as HTMLInputElement;
    this.selectedFile = input.files?.[0] ?? null;
    this.importError.set('');
  }

  startWakaTimeImport() {
    if (!this.wakaTimeApiKey.trim()) {
      this.importError.set('Enter your WakaTime API key, or upload an export file instead.');
      return;
    }

    this.importing.set(true);
    this.importStarted.set(false);
    this.importError.set('');

    this.api.post('/api/import/wakatime', { apiKey: this.wakaTimeApiKey.trim() }).subscribe({
      next: () => {
        this.importing.set(false);
        this.importStarted.set(true);
      },
      error: (err) => {
        this.importing.set(false);
        this.importError.set(err.error?.message ?? 'Failed to start WakaTime import');
      },
    });
  }

  importFromFile() {
    if (!this.selectedFile) {
      this.importError.set('Choose an export file first.');
      return;
    }

    const formData = new FormData();
    formData.append('file', this.selectedFile);

    this.importing.set(true);
    this.importStarted.set(true);
    this.importError.set('');
    this.importProgress.set(null);

    this.api.postFormData('/api/import/file', formData).subscribe({
      next: () => this.pollImportStatus(),
      error: (err) => {
        this.importing.set(false);
        this.importStarted.set(false);
        this.importError.set(err.error?.message ?? 'Import failed');
      },
    });
  }

  importProgressPercent(): number {
    const progress = this.importProgress();
    if (!progress?.total) return 0;
    return Math.min(100, (progress.imported / progress.total) * 100);
  }

  finishOnboarding() {
    this.stopPolling();
    void this.router.navigate(['/dashboard']);
  }

  private pollImportStatus() {
    this.stopPolling();

    this.importPollInterval = setInterval(() => {
      this.api.get<ImportProgress>('/api/import/status').subscribe({
        next: (status) => {
          this.importProgress.set(status);

          if (status.status === 'completed') {
            this.stopPolling();
            this.importing.set(false);
          } else if (status.status === 'failed') {
            this.stopPolling();
            this.importing.set(false);
            this.importError.set(status.error || 'Import failed');
          }
        },
        error: () => {
          this.stopPolling();
          this.importing.set(false);
          this.importError.set('Could not check import progress');
        },
      });
    }, 1500);
  }

  private stopPolling() {
    if (this.importPollInterval) {
      clearInterval(this.importPollInterval);
      this.importPollInterval = null;
    }
  }
}
