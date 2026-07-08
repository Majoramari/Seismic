import { Component, inject, OnInit, ChangeDetectorRef, signal } from '@angular/core';
import { FormsModule } from '@angular/forms';
import { ToastService } from '../../core/toast/toast.service';
import {
  LucideAngularModule,
  Clock,
  Calendar,
  Flame,
  Star,
  Trophy,
  Users,
  Pencil,
  Camera,
  Trash2,
  Info,
  Eye,
  FileText,
  MapPin,
  UsersRound,
  Languages,
  Check,
  Plus,
  Flag,
  Code,
  Crown,
  Sun,
  Zap,
  Award,
  FolderOpen,
  Activity,
  CircleUser,
  Link as LinkIcon,
} from 'lucide-angular';
import { ProfileService, ProfileResponse, HeatmapCell, Achievement, ActivityLogItem, UpdateProfileRequest } from './profile.service';
import { MagicBentoDirective, MagicBentoCardDirective } from '../../shared/components/magic-bento/magic-bento.directive';
// ── Local display interfaces (icons are frontend-only) ────────────────────────

interface Metric {
  icon: typeof Clock;
  iconClass: string;
  label: string;
  value: string;
  sub: string;
  positive: boolean;
}

interface DisplayAchievement {
  icon: typeof Sun;
  badgeClass: string;
  title: string;
  description: string;
  date: string;
}

interface DisplayActivity {
  icon: typeof Zap;
  text: string;
  time: string;
}

interface DisplayInfoField {
  icon: typeof CircleUser;
  label: string;
  completed: boolean;
}

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [LucideAngularModule, FormsModule, MagicBentoDirective, MagicBentoCardDirective],
  templateUrl: './profile.html',
  styleUrl: './profile.css',
})
export class Profile implements OnInit {
  private profileService = inject(ProfileService);
  private cdr = inject(ChangeDetectorRef);
  private toast = inject(ToastService);

  // ── Icons (static, not from backend) ─────────────────────────────────────
  readonly ClockIcon = Clock;
  readonly CalendarIcon = Calendar;
  readonly FlameIcon = Flame;
  readonly StarIcon = Star;
  readonly TrophyIcon = Trophy;
  readonly UsersIcon = Users;
  readonly PencilIcon = Pencil;
  readonly CameraIcon = Camera;
  readonly TrashIcon = Trash2;
  readonly InfoIcon = Info;
  readonly EyeIcon = Eye;
  readonly FileTextIcon = FileText;
  readonly MapPinIcon = MapPin;
  readonly UsersRoundIcon = UsersRound;
  readonly LanguagesIcon = Languages;
  readonly CheckIcon = Check;
  readonly PlusIcon = Plus;
  readonly FlagIcon = Flag;
  readonly CodeIcon = Code;
  readonly CrownIcon = Crown;
  readonly SunIcon = Sun;
  readonly ZapIcon = Zap;
  readonly AwardIcon = Award;
  readonly FolderIcon = FolderOpen;
  readonly ActivityIcon = Activity;
  readonly CircleUserIcon = CircleUser;
  readonly LinkIcon = LinkIcon;

  // ── State ─────────────────────────────────────────────────────────────────
  loading = true;
  error: string | null = null;
  profileData: ProfileResponse | null = null;

  showEditModal = signal(false);
  savingProfile = signal(false);
  editForm: UpdateProfileRequest = {};
  editLanguagesStr = '';

  // ── Profile fields (populated from API) ───────────────────────────────────
  username = '';
  firstName = '';
  email = '';
  role = '';
  location = '';
  university = '';
  bio = '';
  joinDate = '';
  lastActive = '';
  timeZone = '';
  memberFor = '';

  // Heatmap (52 weeks × 7 days from backend)
  heatmapData: HeatmapCell[][] = [];
  totalActiveDays = 0;
  maxStreak = 0;

  // Problem Solving gauge
  solved = 0;
  totalProblems = 0;
  attempting = 0;

  // Gauge arc paths (computed from API data)
  gaugeArcs = this.generateGaugeArcs();

  // Achievements
  achievements: DisplayAchievement[] = [];

  // Recent Activity
  recentActivity: DisplayActivity[] = [];

  // Personal info completion
  completionPercent = 0;
  infoFields: DisplayInfoField[] = [];

  // Metrics row (built from API numeric fields)
  metrics: Metric[] = [];

  // Static labels for the heatmap axes
  readonly months = ['Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec', 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'];
  readonly dayLabels = ['Mon', '', 'Wed', '', 'Fri', '', ''];

  // ── Lifecycle ─────────────────────────────────────────────────────────────
  ngOnInit(): void {
    console.log('ngOnInit: calling profileService.getProfile()');
    this.profileService.getProfile().subscribe({
      next: (data) => {
        console.log('API response received data:', data);
        try {
          this.applyProfile(data);
        } catch (e) {
          console.error('Error applying profile:', e);
          this.error = 'Failed to apply profile data';
          this.loading = false;
        }
        this.cdr.detectChanges();
      },
      error: (err) => {
        console.error('API request error:', err);
        this.error = err?.error?.message ?? 'Failed to load profile. Please try again.';
        this.loading = false;
        this.cdr.detectChanges();
      },
    });
  }

  // ── Progress ring (reactive getters) ─────────────────────────────────────
  get progressCircumference(): number {
    return 2 * Math.PI * 18; // radius = 18
  }

  get progressOffset(): number {
    return this.progressCircumference - (this.completionPercent / 100) * this.progressCircumference;
  }

  // ── Private helpers ───────────────────────────────────────────────────────

  private applyProfile(data: ProfileResponse): void {
    this.profileData = data;
    // Identity
    this.username = data.username;
    this.firstName = data.firstName || data.username;
    this.email = data.email;
    this.role = data.role || 'Developer';
    this.location = data.location || '—';
    this.university = data.university || '—';
    this.bio = data.bio || '';
    this.joinDate = data.joinDate;
    this.lastActive = data.lastActive;
    this.timeZone = data.timeZone || '—';
    this.memberFor = data.memberFor;

    // Metrics row
    this.totalActiveDays = data.totalActiveDays;
    this.maxStreak = data.maxStreak;
    this.metrics = this.buildMetrics(data);

    // Problem solving
    this.solved = data.solved;
    this.totalProblems = data.totalProblems;
    this.attempting = data.attempting;
    this.gaugeArcs = this.generateGaugeArcs();

    // Heatmap
    this.heatmapData = data.heatmap ?? [];

    // Achievements
    this.achievements = (data.achievements ?? []).map((a) => this.mapAchievement(a));

    // Recent activity
    this.recentActivity = (data.recentActivity ?? []).map((item) => this.mapActivity(item));

    // Personal info
    this.completionPercent = data.completionPercent;
    this.infoFields = (data.infoFields ?? []).map((f) => ({
      icon: this.iconForInfoKey(f.key),
      label: f.label,
      completed: f.completed,
    }));

    this.loading = false;
  }

  private buildMetrics(data: ProfileResponse): Metric[] {
    const { totalCodingSeconds, totalActiveDays, currentStreak, maxStreak } = data;

    const totalHours = Math.floor(totalCodingSeconds / 3600);
    const totalMins = Math.floor((totalCodingSeconds % 3600) / 60);
    const codingTimeValue = totalCodingSeconds > 0 ? `${totalHours}h ${totalMins}m` : '0h 0m';

    return [
      {
        icon: Clock,
        iconClass: 'green',
        label: 'Total Coding Time',
        value: codingTimeValue,
        sub: 'All time',
        positive: true,
      },
      {
        icon: Calendar,
        iconClass: 'blue',
        label: 'Active Days',
        value: String(totalActiveDays),
        sub: 'Days with activity',
        positive: true,
      },
      {
        icon: Flame,
        iconClass: 'purple',
        label: 'Current Streak',
        value: `${currentStreak} days`,
        sub: `Best: ${maxStreak} days`,
        positive: false,
      },
      {
        icon: Star,
        iconClass: 'gold',
        label: 'Contest Rating',
        // TODO: no data source yet — user_problem_stats is a placeholder
        value: String(data.solved > 0 ? 0 : 0),
        sub: 'Placeholder',
        positive: false,
      },
      {
        icon: Trophy,
        iconClass: 'blue',
        label: 'Contribution',
        // TODO: no data source yet — user_problem_stats is a placeholder
        value: '—',
        sub: 'Placeholder',
        positive: false,
      },
      {
        icon: Users,
        iconClass: 'orange',
        label: 'Friends',
        // TODO: no data source yet
        value: '—',
        sub: 'No friends data',
        positive: false,
      },
    ];
  }

  /** Maps a backend Achievement to a display object with a frontend icon. */
  private mapAchievement(a: Achievement): DisplayAchievement {
    const iconMap: Record<string, typeof Sun> = {
      first_blood: Sun,
      night_owl: Star,
      streak_7: Flame,
      default: Award,
    };
    return {
      icon: iconMap[a.key] ?? iconMap['default'],
      badgeClass: a.badgeClass,
      title: a.title,
      description: a.description,
      date: `Earned ${a.earnedAt}`,
    };
  }

  /** Maps a backend ActivityLogItem to a display object with a frontend icon. */
  private mapActivity(item: ActivityLogItem): DisplayActivity {
    const iconMap: Record<string, typeof Zap> = {
      achievement: Award,
      streak: Flame,
      profile: CircleUser,
      project: FolderOpen,
      default: Zap,
    };
    return {
      icon: iconMap[item.kind] ?? iconMap['default'],
      text: item.text,
      time: this.formatRelativeTime(item.at),
    };
  }

  /** Returns the correct icon for a personal-info field key. */
  private iconForInfoKey(key: string): typeof CircleUser {
    const map: Record<string, typeof CircleUser> = {
      fullName: CircleUser,
      bio: FileText,
      location: MapPin,
      gender: UsersRound,
      languages: Languages,
    };
    return map[key] ?? CircleUser;
  }

  /** Converts an RFC3339 timestamp to a short relative-time string. */
  private formatRelativeTime(at: string): string {
    if (!at) return '';
    const diff = Date.now() - new Date(at).getTime();
    const minutes = Math.floor(diff / 60_000);
    if (minutes < 1) return 'just now';
    if (minutes < 60) return `${minutes} minute${minutes > 1 ? 's' : ''} ago`;
    const hours = Math.floor(minutes / 60);
    if (hours < 24) return `${hours} hour${hours > 1 ? 's' : ''} ago`;
    const days = Math.floor(hours / 24);
    if (days < 7) return `${days} day${days > 1 ? 's' : ''} ago`;
    const weeks = Math.floor(days / 7);
    return `${weeks} week${weeks > 1 ? 's' : ''} ago`;
  }

  /** Recomputes the gauge SVG arc paths whenever solved/totalProblems change. */
  private generateGaugeArcs(): { tealPath: string; goldPath: string; redPath: string } {
    const cx = 100, cy = 100, r = 80;
    const startAngle = Math.PI;

    const tealEnd = startAngle + Math.PI * 0.6;
    const goldEnd = startAngle + Math.PI * 0.85;
    const redEnd = startAngle + Math.PI * 1.0;

    const arc = (start: number, end: number): string => {
      const x1 = cx + r * Math.cos(start);
      const y1 = cy + r * Math.sin(start);
      const x2 = cx + r * Math.cos(end);
      const y2 = cy + r * Math.sin(end);
      const large = end - start > Math.PI ? 1 : 0;
      return `M ${x1} ${y1} A ${r} ${r} 0 ${large} 1 ${x2} ${y2}`;
    };

    return {
      tealPath: arc(startAngle, tealEnd),
      goldPath: arc(tealEnd, goldEnd),
      redPath: arc(goldEnd, redEnd),
    };
  }

  // ── Edit Profile Modal ───────────────────────────────────────────────────

  openEditModal() {
    if (!this.profileData) return;
    this.editForm = {
      firstName: this.profileData.firstName || '',
      bio: this.profileData.bio || '',
      location: this.profileData.location || '',
      role: this.profileData.role || '',
      university: this.profileData.university || '',
      website: this.profileData.website || '',
      gender: this.profileData.gender || '',
      timeZone: this.profileData.timeZone || '',
    };
    this.editLanguagesStr = (this.profileData.languages || []).join(', ');
    this.showEditModal.set(true);
  }

  closeEditModal() {
    this.showEditModal.set(false);
  }

  saveProfile() {
    const langs = this.editLanguagesStr
      .split(',')
      .map((s) => s.trim())
      .filter((s) => s.length > 0);
      
    this.editForm.languages = langs;

    this.savingProfile.set(true);
    this.profileService.updateProfile(this.editForm).subscribe({
      next: () => {
        this.savingProfile.set(false);
        this.showEditModal.set(false);
        this.toast.success('Profile updated successfully');
        this.loading = true;
        this.ngOnInit();
      },
      error: (err) => {
        this.savingProfile.set(false);
        this.toast.error(err?.error?.message ?? 'Failed to update profile');
      },
    });
  }
}
