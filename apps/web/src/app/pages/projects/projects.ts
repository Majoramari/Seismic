import { Component, computed, HostListener, inject, OnInit, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import {
  Archive,
  ArchiveRestore,
  Clock3,
  Code2,
  ExternalLink,
  FolderGit2,
  GitCommitHorizontal,
  Github,
  LucideAngularModule,
} from 'lucide-angular';
import { retry, timer } from 'rxjs';
import { ApiService } from '../../core/api/api.service';
import { ToastService } from '../../core/toast/toast.service';

interface ProjectLanguage {
  language: string;
  seconds: number;
}

interface ProjectOverview {
  project: string;
  seconds: number;
  repoUrl: string | null;
  websiteUrl: string | null;
  lastWorkedAt: string | null;
  lastCommitAt: string | null;
  languages: ProjectLanguage[];
  archived: boolean;
}

type RangeOption = 'today' | 'week' | 'month' | 'all';
type ProjectTab = 'active' | 'archived';

const LABEL_OVERRIDES = new Map<string, string>([
  ['css', 'CSS'],
  ['html', 'HTML'],
  ['javascript', 'JavaScript'],
  ['typescript', 'TypeScript'],
  ['json', 'JSON'],
  ['jsx', 'JSX'],
  ['tsx', 'TSX'],
  ['yaml', 'YAML'],
  ['yml', 'YAML'],
  ['sql', 'SQL'],
  ['go', 'Go'],
  ['golang', 'Go'],
]);
const PROJECT_PAGE_SIZE = 6;

@Component({
  selector: 'app-projects',
  standalone: true,
  imports: [FormsModule, LucideAngularModule],
  templateUrl: './projects.html',
})
export class Projects implements OnInit {
  private api = inject(ApiService);
  private toast = inject(ToastService);

  readonly FolderGitIcon = FolderGit2;
  readonly GithubIcon = Github;
  readonly ExternalLinkIcon = ExternalLink;
  readonly ArchiveIcon = Archive;
  readonly ArchiveRestoreIcon = ArchiveRestore;
  readonly ClockIcon = Clock3;
  readonly CodeIcon = Code2;
  readonly CommitIcon = GitCommitHorizontal;

  range = signal<RangeOption>('all');
  tab = signal<ProjectTab>('active');
  loading = signal(true);
  loadingMore = signal(false);
  hasMore = signal(true);
  projects = signal<ProjectOverview[]>([]);
  updating = signal<string | null>(null);

  totalSeconds = computed(() =>
    this.projects().reduce((total, project) => total + project.seconds, 0),
  );

  ngOnInit() {
    this.loadProjects();
  }

  setTab(tab: ProjectTab) {
    if (this.tab() === tab) return;
    this.tab.set(tab);
    this.loadProjects(true);
  }

  setRange(range: RangeOption) {
    this.range.set(range);
    this.loadProjects(true);
  }

  @HostListener('window:scroll')
  onWindowScroll() {
    const scrollPosition = window.scrollY + window.innerHeight;
    const threshold = document.documentElement.scrollHeight - 600;
    if (scrollPosition >= threshold) this.loadProjects(false);
  }

  toggleArchive(project: ProjectOverview) {
    this.updating.set(project.project);
    this.api
      .post<null>('/api/projects/archive', {
        project: project.project,
        archived: !project.archived,
      })
      .subscribe({
        next: () => {
          this.projects.update((items) => items.filter((item) => item.project !== project.project));
          this.toast.success(project.archived ? 'Project restored' : 'Project archived');
          this.updating.set(null);
        },
        error: () => {
          this.toast.error('Could not update project');
          this.updating.set(null);
        },
      });
  }

  projectName(project: ProjectOverview): string {
    return project.project.trim() || 'Unnamed project';
  }

  loadProjects(reset = true) {
    if (!reset && (!this.hasMore() || this.loading() || this.loadingMore())) return;

    if (reset) {
      this.projects.set([]);
      this.hasMore.set(true);
      this.loading.set(true);
    } else {
      this.loadingMore.set(true);
    }

    this.api
      .get<ProjectOverview[]>('/api/projects', {
        range: this.range(),
        archived: this.tab() === 'archived' ? 'true' : 'false',
        limit: String(PROJECT_PAGE_SIZE),
        offset: String(reset ? 0 : this.projects().length),
      })
      .pipe(retry({ count: 2, delay: (_, retryIndex) => timer(retryIndex * 500) }))
      .subscribe({
        next: (projects) => {
          const nextProjects = projects ?? [];
          this.projects.set(reset ? nextProjects : [...this.projects(), ...nextProjects]);
          this.hasMore.set(nextProjects.length === PROJECT_PAGE_SIZE);
          this.loading.set(false);
          this.loadingMore.set(false);
          this.loadMoreIfPageDoesNotScroll();
        },
        error: () => {
          if (reset) this.projects.set([]);
          this.loading.set(false);
          this.loadingMore.set(false);
          this.toast.error('Could not load projects');
        },
      });
  }

  private loadMoreIfPageDoesNotScroll() {
    window.setTimeout(() => {
      if (!this.hasMore() || this.loading() || this.loadingMore()) return;
      const pageDoesNotScroll = document.documentElement.scrollHeight <= window.innerHeight + 80;
      if (pageDoesNotScroll) this.loadProjects(false);
    });
  }

  formatSeconds(seconds: number): string {
    if (seconds < 60) return '< 1m';
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    const remainingSeconds = seconds % 60;
    const parts: string[] = [];

    if (hours > 0) parts.push(`${hours}h`);
    if (minutes > 0) parts.push(`${minutes}m`);
    if (hours === 0 && remainingSeconds > 0) parts.push(`${remainingSeconds}s`);

    return parts.join(' ');
  }

  relativeTime(value: string | null): string {
    if (!value) return 'Not recorded yet';

    const date = new Date(value);
    const diffMs = Date.now() - date.getTime();
    if (Number.isNaN(diffMs)) return 'Not recorded yet';

    const minute = 60 * 1000;
    const hour = 60 * minute;
    const day = 24 * hour;
    const month = 30 * day;

    if (diffMs < minute) return 'just now';
    if (diffMs < hour) return `${Math.floor(diffMs / minute)}m ago`;
    if (diffMs < day) return `${Math.floor(diffMs / hour)}h ago`;
    if (diffMs < month) return `${Math.floor(diffMs / day)}d ago`;

    return date.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
  }

  languageSummary(project: ProjectOverview): string {
    if (project.languages.length === 0) return 'No languages recorded';

    return project.languages
      .slice(0, 4)
      .map((language) => this.formatDisplayLabel(language.language))
      .join(', ');
  }

  formatDisplayLabel(value: string | null | undefined): string {
    if (!value) return '-';

    const trimmed = value.trim();
    if (!trimmed) return '-';

    const override = LABEL_OVERRIDES.get(trimmed.toLowerCase());
    if (override) return override;

    return trimmed
      .split(/([\s._-]+)/)
      .map((part) => {
        if (/^[\s._-]+$/.test(part)) return part;
        return part.charAt(0).toUpperCase() + part.slice(1);
      })
      .join('');
  }
}
