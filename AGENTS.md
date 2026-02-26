# Ovumcy Project Agent Instructions (Enterprise-level)

## Project Context

- Project: ovumcy — privacy-critical application that works with sensitive health-related data.
- Goal: clean, maintainable Go backend + web frontend with strong privacy and security guarantees.
- You must treat this repository as production-grade, not as a playground.

## Architecture and Layers

- Backend: Go.
  - Entrypoint in `cmd/`.
  - HTTP/transport layer in `internal/api`.
  - Business logic in `internal/services`.
  - Persistence / DB access in `internal/db`.
  - Domain models in `internal/models`.
  - Cross-cutting concerns in `internal/security`, `internal/i18n`, `internal/templates`.
- Frontend: `web/` (Tailwind CSS, JS, templates).
- DB migrations: `migrations/` — changes only on explicit request with a written plan.
- Local data: `.local/`, `data/` — considered sensitive; never touch without explicit permission.

## Strict Anti-Patterns (No Tech Debt, No Vibe Coding)

- Do not:
  - mix transport, business logic, and persistence in the same function or file;
  - add “just one more if” into already complex handlers instead of extracting helpers/services;
  - duplicate validation, mapping, or business rules across handlers/services;
  - add TODOs without owner, reason, and issue reference;
  - introduce temporary hacks “until later” without explicitly labeling them as debt and getting approval.
- Any change that would increase complexity, coupling, or duplication must be rejected or refactored into a cleaner design.

## Decomposition and Planning

- For any non-trivial change (more than one file or function):
  - produce a numbered plan of small, atomic steps;
  - explicitly map each step to a layer (`api`, `services`, `db`, `models`, `web`);
  - highlight which steps touch privacy/security invariants;
  - wait for approval before editing files.
- If you discover new complexity mid-way, stop, update the plan, and request approval again.

## Privacy and Security Invariants

- Treat all user-related and health-related data as highly sensitive.
- Partner/viewer roles must never see private owner-only fields.
- All write operations must enforce:
  - authenticated user,
  - correct role,
  - valid CSRF and authorization context.
- Never weaken or bypass privacy checks for convenience.
- Any change touching `security`, auth/session handling, access control, logging of PII, or export flows is security-sensitive:
  - call it out explicitly,
  - propose additional review and tests.

## Testing and Observability

- After backend changes:
  - propose `go test ./...` from project root (or a more targeted package when appropriate).
- After frontend changes:
  - propose `npm run lint` and `npm run build` before considering the change safe.
- For changes in critical flows (auth, permissions, data export, cycle/health logic):
  - recommend adding or updating tests that cover:
    - happy path,
    - edge cases,
    - unauthorized/forbidden access.
- Prefer adding minimal but meaningful logging/metrics in critical paths rather than leaving them opaque.

## Deployment and Environment Safety

- Never modify `docker-compose.yml`, `docker/`, `migrations/`, `.local/`, `data/`, or environment-related files without explicit approval and a rollback plan.
- Do not introduce new external dependencies (libraries, services, APIs) without:
  - explaining why they are needed,
  - checking license / operational impact,
  - getting explicit approval.

## Governance

- You may propose improvements to these project rules and point out missing invariants or recurring issues.
- You must not modify `AGENTS.md`, `SKILL.md`, or other governance files yourself unless the user explicitly requests a concrete change and approves the diff.

## Backend layering rules (ovumcy)

- `internal/api` must not access the database directly. All persistence goes through repositories in `internal/db`.
- Business logic (cycle calculations, onboarding, settings, symptoms, etc.) must live in `internal/services` with unit tests, not in handlers or templates.
- `internal/db` is responsible only for persistence (CRUD, queries), with no business decisions or HTTP concerns.
- New features must follow this layering from day one: API → services → repositories → models.
- `internal/api` owns only transport concerns: request parsing, auth/role checks, input validation, and HTTP response mapping.
- Service layer must not depend on Fiber (`*fiber.Ctx`) or HTTP response codes; it returns domain data/errors.
- Repository methods should expose narrow operations (no handler-side query construction), so persistence details stay in `internal/db`.
- Handler wiring should use dependency injection (`repositories` + `services`); direct DB fallback is allowed only for backward-compatible test setup.
