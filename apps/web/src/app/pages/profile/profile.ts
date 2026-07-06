import { Component } from '@angular/core';
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

interface HeatmapCell {
  level: number;
}

interface Achievement {
  icon: typeof Sun;
  badgeClass: string;
  title: string;
  description: string;
  date: string;
}

interface ActivityLogItem {
  icon: typeof Zap;
  text: string;
  time: string;
}

interface Metric {
  icon: typeof Clock;
  iconClass: string;
  label: string;
  value: string;
  sub: string;
  positive: boolean;
}

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [LucideAngularModule],
  templateUrl: './profile.html',
  styleUrl: './profile.css',
})
export class Profile {
  // Icons
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

  // Profile data
  readonly username = 'Amor_Mousa';
  readonly firstName = 'Amr';
  readonly email = 'amor.mousa444@gmail.com';
  readonly role = 'Frontend Developer';
  readonly location = 'Cairo, Egypt';
  readonly university = 'MTI University';
  readonly bio = 'Building things, breaking things, learning from everything.';
  readonly joinDate = 'May 2023';
  readonly lastActive = 'Online now';
  readonly timeZone = 'Africa/Cairo';
  readonly memberFor = '2 years';

  // Metrics
  readonly metrics: Metric[] = [
    {
      icon: Clock,
      iconClass: 'green',
      label: 'Total Coding Time',
      value: '3985h 24m',
      sub: '+120h this week',
      positive: true,
    },
    {
      icon: Calendar,
      iconClass: 'blue',
      label: 'Active Days',
      value: '247',
      sub: '+5 this week',
      positive: true,
    },
    {
      icon: Flame,
      iconClass: 'purple',
      label: 'Current Streak',
      value: '24 days',
      sub: 'Best: 67 days',
      positive: false,
    },
    {
      icon: Star,
      iconClass: 'gold',
      label: 'Contest Rating',
      value: '1216',
      sub: 'Max: 1242',
      positive: false,
    },
    {
      icon: Trophy,
      iconClass: 'blue',
      label: 'Contribution',
      value: '-2',
      sub: 'Global Rank',
      positive: false,
    },
    {
      icon: Users,
      iconClass: 'orange',
      label: 'Friends',
      value: '136',
      sub: 'Connected',
      positive: false,
    },
  ];

  // Heatmap data — generate a full year grid (52 weeks × 7 days)
  readonly months = ['Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec', 'Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun'];
  readonly dayLabels = ['Mon', '', 'Wed', '', 'Fri', '', ''];
  readonly heatmapData: HeatmapCell[][] = this.generateHeatmap();
  readonly totalActiveDays = 186;
  readonly maxStreak = 67;

  // Problem Solving gauge
  readonly solved = 0;
  readonly totalProblems = 3985;
  readonly attempting = 0;

  // Gauge arc paths
  readonly gaugeArcs = this.generateGaugeArcs();

  // Achievements
  readonly achievements: Achievement[] = [
    {
      icon: Sun,
      badgeClass: 'gold',
      title: 'Early Bird',
      description: 'Solved a problem before 9 AM',
      date: 'Earned May 20, 2024',
    },
    {
      icon: Crown,
      badgeClass: 'yellow',
      title: 'Consistency King',
      description: 'Active for 30 days in a row',
      date: 'Earned Mar 15, 2024',
    },
    {
      icon: Code,
      badgeClass: 'teal',
      title: 'Problem Solver',
      description: 'Solved 100 problems',
      date: 'Earned Jan 10, 2024',
    },
    {
      icon: Flag,
      badgeClass: 'purple',
      title: 'First Submission',
      description: 'Made your first submission',
      date: 'Earned May 1, 2023',
    },
  ];

  // Recent Activity
  readonly recentActivity: ActivityLogItem[] = [
    { icon: Zap, text: 'Solved 3 problems', time: '2 hours ago' },
    { icon: Activity, text: 'Active coding session', time: '3 hours ago' },
    { icon: Award, text: 'Earned "Early Bird" badge', time: '1 day ago' },
    { icon: CircleUser, text: 'Updated profile information', time: '3 days ago' },
    { icon: FolderOpen, text: 'Added new project', time: '1 week ago' },
  ];

  // Personal info completion
  readonly completionPercent = 5;
  readonly infoFields = [
    { label: 'Full name', completed: true, icon: CircleUser },
    { label: 'Bio', completed: false, icon: FileText },
    { label: 'Location', completed: false, icon: MapPin },
    { label: 'Gender', completed: false, icon: UsersRound },
    { label: 'Languages', completed: false, icon: Languages },
  ];

  // Progress ring calculations
  get progressCircumference(): number {
    return 2 * Math.PI * 18; // radius = 18
  }

  get progressOffset(): number {
    return this.progressCircumference - (this.completionPercent / 100) * this.progressCircumference;
  }

  private generateHeatmap(): HeatmapCell[][] {
    const weeks: HeatmapCell[][] = [];

    // Build 52 weeks. Add some realistic activity in later months.
    for (let w = 0; w < 52; w++) {
      const week: HeatmapCell[] = [];
      for (let d = 0; d < 7; d++) {
        let level = 0;
        // Add activity primarily in weeks 26-48 (Jan–May in the chart)
        if (w >= 26 && w <= 48) {
          const rand = Math.random();
          if (rand > 0.7) level = 4;
          else if (rand > 0.5) level = 3;
          else if (rand > 0.3) level = 2;
          else if (rand > 0.15) level = 1;
        } else if (w >= 20 && w < 26) {
          // Some sparse activity
          const rand = Math.random();
          if (rand > 0.85) level = 2;
          else if (rand > 0.7) level = 1;
        }
        week.push({ level });
      }
      weeks.push(week);
    }

    return weeks;
  }

  private generateGaugeArcs(): { tealPath: string; goldPath: string; redPath: string } {
    // Semi-circular gauge from 180° to 360° (bottom-up)
    const cx = 100, cy = 100, r = 80;
    const startAngle = Math.PI; // 180°

    // Teal: 0-60%  Gold: 60-85%  Red: 85-100%
    const tealEnd = startAngle + Math.PI * 0.6;
    const goldEnd = startAngle + Math.PI * 0.85;
    const redEnd = startAngle + Math.PI * 1.0;

    const arc = (start: number, end: number) => {
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
}
