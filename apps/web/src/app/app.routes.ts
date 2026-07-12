import { Routes } from '@angular/router';
import { Login } from './pages/auth/login/login';
import { Verify } from './pages/auth/verify/verify';
import { Dashboard } from './pages/dashboard/dashboard';
import { Leaderboard } from './pages/leaderboard/leaderboard';
import { authGuard } from './core/auth/auth.guard';
import { guestGuard } from './core/auth/guest.guard';
import { Settings } from './pages/settings/settings';
import { Landing } from './pages/landing/landing';

export const routes: Routes = [
  { path: '', component: Landing, canActivate: [guestGuard], title: 'Seismic — Track your coding activity' },
  { path: 'login', component: Login, canActivate: [guestGuard], title: 'Log in — Seismic' },
  { path: 'verify', component: Verify, title: 'Verifying — Seismic' },
  {
    path: 'dashboard',
    component: Dashboard,
    canActivate: [authGuard],
    title: 'Dashboard — Seismic',
  },
  { path: 'leaderboard', component: Leaderboard, title: 'Leaderboard — Seismic' },
  { path: 'settings', component: Settings, canActivate: [authGuard], title: 'Settings — Seismic' },
];
