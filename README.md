# Seismic

Seismic is a coding activity tracker for developers who want a private, cross-editor view of where their time goes. It records activity from your editor, turns that activity into useful stats, and shows the results in a web dashboard.

It works across VS Code, JetBrains IDEs, and Neovim, so your stats stay unified even if you move between editors during the day.

## Project Links

- Live app: [https://seismic.icu](https://seismic.icu)
- Contact: [hello@seismic.icu](mailto:hello@seismic.icu)

## Features

- Coding time by day, week, month, year, and all time.
- Project, language, editor, operating system, and machine breakdowns.
- A dashboard with summary cards, charts, heatmap activity, goals, and a project timeline.
- Public profile pages with privacy controls for hiding sensitive stats.
- Goals and reminders for daily, weekly, or monthly coding targets.
- Badges for milestones and profile achievements.
- WakaTime import support so existing history can be brought into Seismic.
- Offline-friendly editor tracking: extensions queue failed heartbeats and retry later.
- Shared API key support across supported editors.
- Optional project metadata sync, including Git repository links, website links, branch names, and recent commits.

## Editor Extensions

Install the editor integration you use, then open your Seismic settings page and copy your API key.

| Editor | Install |
| --- | --- |
| VS Code | [Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=muhannad.seismic-stats) |
| JetBrains IDEs | [JetBrains Marketplace search](https://plugins.jetbrains.com/search?search=Seismic) or build from `apps/jetbrains` |
| Neovim | See [apps/nvim/README.md](apps/nvim/README.md) |

After installation:

1. Sign in to the Seismic web app.
2. Go to Settings.
3. Copy your API key.
4. Run the editor command to set the API key:
   - VS Code: `Seismic: Set API Key`
   - JetBrains: `Tools > Seismic > Set API Key`
   - Neovim: `:SeismicSetApiKey`

## Self-Hosting

Requirements:

- PostgreSQL.
- Go 1.25+ for the API.
- Bun 1.3+ for the web app and VS Code extension build.
- Node.js compatible with Angular 22. The Angular CLI used here requires Node `22.22.3+`, `24.15.0+`, or `26.0.0+`.
- Java 17+ for the JetBrains plugin build.
- `curl` and `git` for the Neovim integration.

Project layout:

```text
apps/api        Fiber API, PostgreSQL migrations, auth, stats, imports, badges, goals
apps/web        Angular web dashboard
apps/vscode     VS Code extension
apps/jetbrains  JetBrains IDE plugin
apps/nvim       Neovim plugin
```

### Local Development

Create an API environment file in `apps/api/.env`:

```env
DATABASE_URL=postgres://user:password@localhost:5432/seismic?sslmode=disable
JWT_SECRET=change-me
PORT=5024
APP_URL=http://localhost:4200
ALLOWED_ORIGINS=http://localhost:4200
SMTP_HOST=
SMTP_PORT=
SMTP_USER=
SMTP_PASS=
CLOUDINARY_CLOUD_NAME=
CLOUDINARY_API_KEY=
CLOUDINARY_API_SECRET=
```

Start the API:

```bash
cd apps/api
go run .
```

The API runs migrations automatically on startup.

Start the web app:

```bash
cd apps/web
bun install
bun run start
```

For local editor testing, point the extension or plugin API URL at `http://localhost:5024`.

### API Deployment

Deploy `apps/api` as a Go service.

Required environment variables:

```env
DATABASE_URL=
JWT_SECRET=
PORT=5024
APP_URL=https://your-web-domain.com
ALLOWED_ORIGINS=https://your-web-domain.com
```

Optional but recommended:

```env
SMTP_HOST=
SMTP_PORT=
SMTP_USER=
SMTP_PASS=
CLOUDINARY_CLOUD_NAME=
CLOUDINARY_API_KEY=
CLOUDINARY_API_SECRET=
```

Run command:

```bash
go run .
```

Build command:

```bash
go build -o seismic-api .
```

The API exposes Swagger docs at `/api/docs/`.

### Web App Deployment

Set the production API URL in `apps/web/src/environments/environment.prod.ts`.

Build:

```bash
cd apps/web
bun install
bun run build
```

Deploy `apps/web/dist/web/browser` as a static site. The included `apps/web/vercel.json` is configured for Vercel with SPA rewrites.

### VS Code Extension

```bash
cd apps/vscode
bun install
bun run build
```

Publish with your VS Code Marketplace publisher account, or package it with `vsce` if you want to distribute a `.vsix`.

### JetBrains Plugin

```bash
cd apps/jetbrains
./gradlew build
```

The plugin artifact is produced under `apps/jetbrains/build/distributions`.

### Neovim Plugin

The Neovim plugin is distributed from this monorepo. Users install `majoramari/seismic` and add `apps/nvim` to the runtime path. See [apps/nvim/README.md](apps/nvim/README.md) for the full setup.

## Privacy

Seismic records coding activity metadata such as file path, project name, language, editor, branch, OS, machine name, line count, cursor line, timezone, and keystroke count. Privacy settings let users hide profile stats, project stats, editor stats, OS stats, badges, and selected projects from public views.

## Contributors

- [Muhannad Elbolaky](https://muhannad.cc)
- Ali Mustafa
- Momen Aymann
- Aya Mohamed
- Amr Mousa
- Basma Bahaa

## License

GPL-3.0. See [LICENSE](LICENSE).
