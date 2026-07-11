# Seismic

Seismic is a developer time tracking platform. It records editor activity from a VS Code extension, sends heartbeat
events to a Go API, and displays coding activity, stats, and leaderboard views in an Angular web app.

## Project Structure

```text
apps/
  api/       Go/Fiber API, PostgreSQL storage, auth, stats, leaderboard, Swagger docs
  web/       Angular dashboard, login, settings, leaderboard, and stats UI
  vscode/    VS Code extension that tracks coding activity and sends heartbeats
bruno/       Bruno API request collections for local API testing
```

## Stack

- API: Go, Fiber, PostgreSQL, pgx, JWT, Swagger
- Web: Angular, Bun, RxJS, lucide-angular

## Requirements

- Go 1.25+
- Bun 1.3+
- PostgreSQL
- VS Code, for extension development
- Bruno, optional for API collection testing

## API Setup

```bash
cd apps/api
cp .env.example .env
```

Run the API:

```bash
go run .
```

The API runs migrations on startup.

## Web Setup

```bash
cd apps/web
bun install
bun run start
```

Development configuration points the web app at:

```text
http://localhost:5024
```

Build the web app:

```bash
bun run build
```
