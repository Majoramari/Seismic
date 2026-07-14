import { Component, computed, inject, OnInit, signal } from '@angular/core';
import { NgOptimizedImage } from '@angular/common';
import { DomSanitizer, SafeHtml } from '@angular/platform-browser';
import { RouterLink } from '@angular/router';
import {
  LucideAngularModule,
  KeyRound,
  Eye,
  EyeOff,
  Copy,
  Check,
  ExternalLink,
} from 'lucide-angular';
import hljs from 'highlight.js/lib/core';
import lua from 'highlight.js/lib/languages/lua';
import { ApiService } from '../../core/api/api.service';
import { AuthService } from '../../core/auth/auth.service';

hljs.registerLanguage('lua', lua);

interface ApiKeyResponse {
  apiKey: string;
}

type Group = 'vscode' | 'jetbrains' | 'neovim';

interface EditorEntry {
  id: string;
  name: string;
  group: Group;
  logo: string; // path under /images/ide-logos/, all rasterized PNG
}

interface PluginLink {
  label: string;
  url: string;
}

@Component({
  selector: 'app-docs',
  standalone: true,
  imports: [RouterLink, LucideAngularModule, NgOptimizedImage],
  templateUrl: './docs.html',
  styleUrl: './docs.css',
})
export class Docs implements OnInit {
  private api = inject(ApiService);
  readonly auth = inject(AuthService);
  private sanitizer = inject(DomSanitizer);

  readonly KeyIcon = KeyRound;
  readonly EyeIcon = Eye;
  readonly EyeOffIcon = EyeOff;
  readonly CopyIcon = Copy;
  readonly CheckIcon = Check;
  readonly ExternalLinkIcon = ExternalLink;

  apiKey = signal('');
  apiKeyVisible = signal(false);
  copied = signal(false);

  readonly editors: EditorEntry[] = [
    { id: 'vscode', name: 'VS Code', group: 'vscode', logo: '/images/ide-logos/vscode.png' },
    {
      id: 'intellij',
      name: 'IntelliJ IDEA',
      group: 'jetbrains',
      logo: '/images/ide-logos/intellij-idea.png',
    },
    { id: 'pycharm', name: 'PyCharm', group: 'jetbrains', logo: '/images/ide-logos/pycharm.png' },
    {
      id: 'webstorm',
      name: 'WebStorm',
      group: 'jetbrains',
      logo: '/images/ide-logos/webstorm.png',
    },
    { id: 'goland', name: 'GoLand', group: 'jetbrains', logo: '/images/ide-logos/goland.png' },
    { id: 'rider', name: 'Rider', group: 'jetbrains', logo: '/images/ide-logos/rider.png' },
    { id: 'clion', name: 'CLion', group: 'jetbrains', logo: '/images/ide-logos/clion.png' },
    {
      id: 'rubymine',
      name: 'RubyMine',
      group: 'jetbrains',
      logo: '/images/ide-logos/rubymine.png',
    },
    {
      id: 'phpstorm',
      name: 'PhpStorm',
      group: 'jetbrains',
      logo: '/images/ide-logos/phpstorm.png',
    },
    {
      id: 'datagrip',
      name: 'DataGrip',
      group: 'jetbrains',
      logo: '/images/ide-logos/datagrip.png',
    },
    {
      id: 'androidstudio',
      name: 'Android Studio',
      group: 'jetbrains',
      logo: '/images/ide-logos/android-studio.png',
    },
    {
      id: 'rustrover',
      name: 'RustRover',
      group: 'jetbrains',
      logo: '/images/ide-logos/rustrover.png',
    },
    { id: 'neovim', name: 'Neovim', group: 'neovim', logo: '/images/ide-logos/neovim.png' },
  ];

  readonly pluginLinks: Record<Group, PluginLink> = {
    vscode: {
      label: 'View on VS Code Marketplace',
      url: 'https://marketplace.visualstudio.com/items?itemName=muhannad.seismic-stats',
    },
    jetbrains: {
      label: 'View on JetBrains Marketplace',
      url: 'https://plugins.jetbrains.com/plugin/32796-seismic/',
    },
    neovim: {
      label: 'View source on GitHub',
      url: 'https://github.com/Majoramari/Seismic/tree/main/apps/nvim',
    },
  };

  readonly neovimLazySpec = `{
  "majoramari/seismic",
  event = { "BufReadPost", "BufNewFile" },
  init = function(plugin)
    vim.opt.rtp:append(plugin.dir .. "/apps/nvim")
  end,
  main = "seismic",
  opts = {},
}`;

  readonly lualineSnippet = `local function get_seismic()
  return require("seismic.statusline").get()
end

return {
  "nvim-lualine/lualine.nvim",
  opts = {
    sections = {
      lualine_x = {
        { get_seismic, icon = "\u{f64e}" }, -- requires a Nerd Font; drop icon if you don't have one
        "encoding",
        "filetype",
      },
    },
  },
}`;

  highlightedLazySpec: SafeHtml = this.highlight(this.neovimLazySpec);
  highlightedLualineSnippet: SafeHtml = this.highlight(this.lualineSnippet);

  private highlight(code: string): SafeHtml {
    const { value } = hljs.highlight(code, { language: 'lua' });
    return this.sanitizer.bypassSecurityTrustHtml(value);
  }

  // Which specific card is selected (not just the group) — fixes the bug
  // where clicking one JetBrains IDE highlighted all of them.
  selected = signal<string>('vscode');

  selectedEditor = computed(
    () => this.editors.find((e) => e.id === this.selected()) ?? this.editors[0],
  );
  selectedGroup = computed<Group>(() => this.selectedEditor().group);
  selectedLink = computed<PluginLink>(() => this.pluginLinks[this.selectedGroup()]);

  ngOnInit() {
    if (!this.auth.isLoggedIn()) return;

    this.api.get<ApiKeyResponse>('/api/auth/apikey').subscribe({
      next: (data) => this.apiKey.set(data.apiKey),
      error: () => this.apiKey.set(''),
    });
  }

  select(id: string) {
    this.selected.set(id);
  }

  toggleVisible() {
    this.apiKeyVisible.set(!this.apiKeyVisible());
  }

  copyKey() {
    navigator.clipboard.writeText(this.apiKey());
    this.copied.set(true);
    setTimeout(() => this.copied.set(false), 1500);
  }
}
