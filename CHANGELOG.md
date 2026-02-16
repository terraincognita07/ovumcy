# Changelog

## 2026-02-16

### Critical bug fixes
- Fixed dashboard navigation reliability by adding explicit `GET /dashboard` route and updating navbar links to point to `/dashboard`.
- Added dedicated registration flow:
  - New public `GET /register` page.
  - Added sign-up link on login page.
  - Updated auth error redirects so registration errors return to `/register`.
- Fixed calendar side editor stale-date bug:
  - Calendar day editor no longer auto-loads a previously selected/old date after month navigation.
  - Editor now shows a "select day" prompt until a day is chosen in the currently visible month.

### Password recovery
- Added password recovery data fields to user model and schema:
  - `recovery_code_hash`
  - `must_change_password`
- Implemented one-time recovery code generation on registration:
  - Format: `LUME-XXXX-XXXX-XXXX`
  - Stored as bcrypt hash in DB.
  - Shown once on a dedicated recovery page with copy/download actions and required "I saved it" confirmation.
- Added full recovery flow:
  - `GET /forgot-password`
  - `POST /api/auth/forgot-password`
  - `GET /reset-password`
  - `POST /api/auth/reset-password`
- Added in-memory recovery attempt rate limiting:
  - Max 5 failed attempts per IP per 15 minutes.
- Added reset-token flow using signed JWT reset tokens.
- Password reset now rotates recovery code and clears `must_change_password`.
- Enforced strong password rules on registration/reset:
  - Minimum 8 chars, at least 1 uppercase, 1 lowercase, and 1 digit.
- Added CLI admin recovery command:
  - `lume reset-password <email>`
  - Sets a temporary password and marks user as required to change password on next login.

### UI/UX improvements
- Improved save UX for HTMX forms:
  - Save buttons now show loading state while request is in flight.
  - Success messages auto-clear after display.
- Added calendar day deletion flow:
  - New owner-only API endpoints:
    - `GET /api/days/:date/exists`
    - `DELETE /api/days/:date`
  - Day editor now shows a red Delete button only when selected day contains data.
  - Delete action asks for confirmation and re-renders the side editor.
  - Calendar grid auto-refreshes after save/delete via `calendar-day-updated` HTMX trigger.
- Improved calendar "today" visibility with a dedicated today pill and stronger highlight.
- Improved statistics UX with safer empty-data handling:
  - Average values render as `-` when unavailable instead of `0.0`.
  - Added guidance note to track 2-3 full cycles for accurate stats.
- Improved accessibility contrast by darkening muted text color.
- Added new localized strings (RU/EN) for auth recovery flow, new routes, save states, and calendar prompts.

### Build/verification
- Rebuilt Tailwind output after style updates (`npm run build:css`).
- Verified Go packages compile and tests pass (`go test ./...`).
