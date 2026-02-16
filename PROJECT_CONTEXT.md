# Lume Project Context

## 1. Snapshot
- Product: privacy-first self-hosted menstrual cycle tracker.
- Backend: Go 1.26, Fiber v2, GORM, SQLite.
- Frontend: server-side templates + HTMX + Alpine + Tailwind.
- Auth: JWT cookie (`lume_auth`) + CSRF cookie/token (`lume_csrf`).
- Locales: `ru` and `en`.

## 2. Current State (2026-02-16)
- Calendar supports day deletion from side editor.
- Day deletion updates both editor panel and month grid via HTMX trigger.
- Today cell has stronger visual highlight (peach border/ring).
- Recovery flow + register flow + dashboard route fixes are already shipped.

## 3. Top-Level Structure
- `cmd/lume/main.go`: runtime bootstrap.
- `internal/api/`: routes, handlers, middleware.
- `internal/models/`: `User`, `SymptomType`, `DailyLog`.
- `internal/services/`: cycle stats/predictions, Telegram reminders.
- `internal/db/sqlite.go`: DB open + AutoMigrate.
- `internal/templates/`: server-rendered pages/partials.
- `internal/i18n/locales/`: `en.json`, `ru.json`.
- `web/src/css/input.css`: Tailwind source.
- `web/static/`: built assets.

## 4. Runtime Flow
1. `cmd/lume/main.go` reads env (`TZ`, `SECRET_KEY`, `DB_PATH`, `PORT`, `DEFAULT_LANGUAGE`).
2. Opens SQLite via `internal/db/OpenSQLite`.
3. Builds i18n manager from `internal/i18n/locales`.
4. Creates API handler with templates.
5. Installs middleware: recover, request logging, compression, language, CSRF.
6. Serves static files from `/static`.
7. Registers routes via `internal/api/RegisterRoutes`.
8. Starts notification loop.

## 5. Route Map

### Public pages
- `GET /healthz`
- `GET /lang/:lang`
- `GET /login`
- `GET /register`
- `GET /forgot-password`
- `GET /reset-password`

### Auth pages
- `GET /`
- `GET /dashboard`
- `GET /calendar`
- `GET /calendar/day/:date` (HTMX day editor panel)
- `GET /stats`

### API
- `/api/auth`
- `GET /setup-status`
- `POST /register`
- `POST /login`
- `POST /forgot-password`
- `POST /reset-password`
- `POST /logout` (auth)

- `/api/days` (auth)
- `GET /` (range by `from`/`to`)
- `GET /:date/exists` (owner-only; returns `{ "exists": bool }`)
- `GET /:date`
- `POST /:date` (owner-only)
- `DELETE /:date` (owner-only)

- `/api/symptoms` (auth)
- `GET /`
- `POST /` (owner-only)
- `DELETE /:id` (owner-only)

- `/api/stats` (auth)
- `GET /overview`

## 6. Calendar/Editor Behavior
- Month grid is rendered in `internal/templates/calendar.html`.
- Side editor partial is `internal/templates/day_editor_partial.html`.
- Grid container `#calendar-grid-panel` listens to `calendar-day-updated from:body`.
- `UpsertDay` and `DeleteDay` set `HX-Trigger: calendar-day-updated`.
- Delete button is shown only when selected day has meaningful data.
- Delete button uses `hx-delete` + `hx-confirm` and re-renders `#day-editor`.

## 7. Data Model
- `User`:
  - `Email` unique
  - `PasswordHash`
  - `RecoveryCodeHash`
  - `MustChangePassword`
  - `Role` (`owner|partner`)
- `SymptomType`:
  - per-user symptom catalog
  - built-in flag + icon/color metadata
- `DailyLog`:
  - unique (`user_id`, `date`)
  - `IsPeriod`, `Flow`, `SymptomIDs` (JSON), `Notes`

## 8. Rules and Safety
- JWT claims include user ID and role.
- `AuthRequired` guards session-protected routes.
- `OwnerOnly` guards all write operations.
- Partner mode is read-only and sanitized (no private notes/symptoms).
- CSRF token is injected in forms and HTMX requests.

## 9. Common Edit Paths
- Route wiring: `internal/api/routes.go`
- Business/request logic: `internal/api/handlers.go`
- Auth/session/language middleware: `internal/api/middleware.go`
- Calendar/day editor UI: `internal/templates/calendar.html`, `internal/templates/day_editor_partial.html`
- Styles: `web/src/css/input.css`
- Locale strings: `internal/i18n/locales/en.json`, `internal/i18n/locales/ru.json`

## 10. Validation Commands
- Run backend checks: `go test ./...`
- Rebuild frontend assets: `npm run build`
- If `npm run build` fails on Windows `cp`, run:
  - `npx tailwindcss -i ./web/src/css/input.css -o ./web/static/css/tailwind.css --minify`
  - `Copy-Item ./node_modules/htmx.org/dist/htmx.min.js ./web/static/js/htmx.min.js -Force`
  - `Copy-Item ./node_modules/alpinejs/dist/cdn.min.js ./web/static/js/alpine.min.js -Force`

## 11. Gotchas
- `cmd/main.go` is deprecated and build-ignored.
- Runtime DB schema follows GORM AutoMigrate, not raw SQL migration files.
- Cookies are `Secure: false` for dev defaults.
- Telegram reminders run only if both `TELEGRAM_BOT_TOKEN` and `TELEGRAM_CHAT_ID` are set.
