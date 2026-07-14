import { Routes } from '@angular/router';
import { Login } from './pages/auth/login/login';
import { Verify } from './pages/auth/verify/verify';
import { Dashboard } from './pages/dashboard/dashboard';
import { Leaderboard } from './pages/leaderboard/leaderboard';
import { Profile } from './pages/profile/profile';
import { Projects } from './pages/projects/projects';
import { authGuard } from './core/auth/auth.guard';
import { guestGuard } from './core/auth/guest.guard';
import { Settings } from './pages/settings/settings';
import { Docs } from './pages/docs/docs';
import { ConfirmEmail } from './pages/confirm-email/confirm-email';

export const routes: Routes = [
  { path: '', redirectTo: 'dashboard', pathMatch: 'full' },
  { path: 'login', component: Login, canActivate: [guestGuard], title: 'Log in — Seismic' },
  { path: 'verify', component: Verify, title: 'Verifying — Seismic' },
  {
    path: 'dashboard',
    component: Dashboard,
    canActivate: [authGuard],
    title: 'Dashboard — Seismic',
  },
  { path: 'leaderboard', component: Leaderboard, title: 'Leaderboard — Seismic' },
  {
    path: 'docs',
    component: Docs,
    canActivate: [authGuard],
    title: 'Docs — Seismic',
  },
  {
    path: 'projects',
    component: Projects,
    canActivate: [authGuard],
    title: 'Projects — Seismic',
  },
  {
    path: 'settings',
    component: Settings,
    canActivate: [authGuard],
    title: 'Settings — Seismic',
  },
  {
    path: 'p/:username',
    component: Profile,
    title: 'Profile — Seismic',
  },
  { path: 'confirm-email', component: ConfirmEmail },
];
