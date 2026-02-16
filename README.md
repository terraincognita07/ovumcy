# Lume

Self-hosted menstrual cycle tracker. Privacy-first, Docker-ready, local-only data ownership.

## Features

- Single-tenant tracking with owner and optional partner role.
- SQLite persistence (`/app/data/lume.db`) with GORM models.
- JWT cookie auth with bcrypt password hashing.
- CSRF protection for all mutating form/API requests.
- Daily log CRUD (period, flow, symptoms, notes).
- Cycle statistics and prediction engine.
- Calendar with actual/predicted period + fertility window highlights.
- Optional Telegram reminders (`TELEGRAM_BOT_TOKEN`, `TELEGRAM_CHAT_ID`).
- HTMX + Alpine templates with local bundled assets.

## Local run

```bash
go run ./cmd/lume
```

Environment variables:

- `PORT` (default `8080`)
- `SECRET_KEY` (default `change_me_in_production`)
- `DB_PATH` (default `data/lume.db`)
- `DEFAULT_LANGUAGE` (default `ru`)
- `TZ` (default `UTC`)
- `TELEGRAM_BOT_TOKEN` (optional)
- `TELEGRAM_CHAT_ID` (optional)

## Docker

From repository root:

```bash
docker compose -f docker/docker-compose.yml up --build
```

App URL: `http://localhost:8080`

## Project structure

- `cmd/lume/main.go`
- `internal/api/`
- `internal/db/`
- `internal/models/`
- `internal/services/`
- `internal/templates/`
- `web/static/`
- `migrations/001_init.sql`
- `docker/Dockerfile`
- `docker/docker-compose.yml`
