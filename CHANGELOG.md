# Changelog

## 2026-02-17

### Onboarding
- Added first-login 3-step onboarding flow for owner accounts:
  - `GET /onboarding`
  - `POST /onboarding/step1`
  - `POST /onboarding/step2`
  - `POST /onboarding/complete`
- Added onboarding gate in auth flow:
  - successful login redirects to `/onboarding` until onboarding is completed;
  - protected pages redirect to onboarding for owner users with incomplete onboarding.
- Extended `User` model and runtime schema (AutoMigrate) with:
  - `onboarding_completed`
  - `cycle_length`
  - `period_length`
  - `last_period_start`
- Added onboarding UI template with Alpine.js + HTMX step transitions and progress bar.
- Added RU/EN locale keys for onboarding content.
- Completing onboarding now creates or updates the first period day record in `daily_logs` using Step 1 date.
- Added onboarding welcome screen and back navigation between onboarding steps.
- Fixed onboarding step HTMX transition state and removed stray `OK` text on successful step save.
- Added cycle stats baseline fallback to onboarding preferences (`cycle_length`, `period_length`, `last_period_start`) while there is not enough real cycle history.
- Added baseline visualization to Stats cycle trend chart:
  - chart now renders onboarding baseline as a dashed line;
  - chart legend now distinguishes actual cycle points from baseline.
- Fixed HTMX save feedback regression for Dashboard/Calendar day forms:
  - successful day save now always shows visible toast feedback;
  - HTMX error responses now render inline in save status blocks;
  - calendar grid refresh is explicitly triggered after successful calendar day save.
- Added onboarding validation tests for date and slider bounds, plus stats-baseline tests for sparse/reliable data cases.

### Data export
- Added owner-only CSV export endpoint:
  - `GET /api/export/csv`
- Added owner-only JSON export endpoint:
  - `GET /api/export/json`
- Added "Export Data" action on Settings page with RU/EN localization.
- Added export note in Settings clarifying scope:
  - only manually tracked entries are exported;
  - predictions (fertile window/ovulation) are not included.
- Added export section summary on Settings:
  - total tracked entries;
  - tracked date range.
- Added toast feedback after successful/failed export download in Settings.

### Data integrity
- Enforced day log validation for period entries:
  - saving a day with `is_period=true` and `flow=none` is now rejected.
- Added localized error message for missing flow when marking a period day.

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
