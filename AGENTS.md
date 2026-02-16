# Lume Agent Notes

## Scope
- Treat this directory (`lume-project/`) as the application root.
- Read `PROJECT_CONTEXT.md` before non-trivial edits.
- Check `CHANGELOG.md` to avoid re-doing already shipped work.

## Entrypoints
- Use `cmd/lume/main.go` as the real runtime entrypoint.
- Ignore `cmd/main.go` for runtime changes (`//go:build ignore`).

## Sources Of Truth
- Runtime schema: GORM models in `internal/models/`.
- DB init + AutoMigrate: `internal/db/sqlite.go`.
- `migrations/001_init.sql` is reference only, not the runtime migration path.

## Fast Validation
- Backend checks: `go test ./...`
- Frontend assets: `npm run build`
- On Windows, `npm run build` may fail on `cp`; use:
  - `npx tailwindcss -i ./web/src/css/input.css -o ./web/static/css/tailwind.css --minify`
  - `Copy-Item ./node_modules/htmx.org/dist/htmx.min.js ./web/static/js/htmx.min.js -Force`
  - `Copy-Item ./node_modules/alpinejs/dist/cdn.min.js ./web/static/js/alpine.min.js -Force`

## Change Routing
- Routes: `internal/api/routes.go`
- Handlers/services glue: `internal/api/handlers.go`
- Auth/language middleware: `internal/api/middleware.go`
- Templates: `internal/templates/`
- i18n locales: `internal/i18n/locales/en.json`, `internal/i18n/locales/ru.json`
- Tailwind source: `web/src/css/input.css`

## Current Calendar Flow (Important)
- Calendar grid lives in `internal/templates/calendar.html`.
- Day editor partial is `internal/templates/day_editor_partial.html`.
- Save and delete actions trigger `HX-Trigger: calendar-day-updated`.
- Calendar grid listens on `calendar-day-updated from:body` and refreshes itself.
- Day delete endpoint: `DELETE /api/days/:date` (owner-only).
- Day data existence endpoint: `GET /api/days/:date/exists` (owner-only).

## Guardrails
- Keep locale keys in `en.json` and `ru.json` synchronized.
- Preserve owner/partner write restrictions (`OwnerOnly` for mutating endpoints).
- Do not introduce migrations as source of truth over AutoMigrate unless explicitly requested.
