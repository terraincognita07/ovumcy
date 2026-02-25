# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Security policy in `SECURITY.md`.
- Contribution guidelines in `CONTRIBUTING.md`.
- Code of conduct in `CODE_OF_CONDUCT.md`.
- Public brand assets (`web/static/brand/*`) and SVG favicon.

### Changed
- Date validation hardened in onboarding and settings:
  - step 1 onboarding date is constrained to an allowed range,
  - settings cycle start date now enforces server-side bounds.
- CI now pins `staticcheck` to a fixed version.
- Docker quick start docs now support a no-clone flow (download docker-compose.yml and .env directly, then run from one folder).
- Docker Compose now uses `pull_policy: always`, so a single `docker compose up -d` pulls the latest image.
- Dashboard cycle-day calculation is now bounded to cycle length (no unbounded growth when the start date is stale).
- Dashboard next period/ovulation predictions are projected into upcoming cycles, and stale-cycle warning text is shown when baseline dates are outdated.
- Mobile calendar day badges/today pill now avoid clipping/overflow with short labels and bounded cell content.
- Date formatting is now locale-aware in dashboard and settings export summaries (RU/EN consistency).
- Settings cycle warnings now render contextually (DOM no longer keeps all warning variants visible at once).
- Stats cards and chart caption now show explicit no-data states instead of misleading default cycle values.
- Settings export range uses native `type="date"` inputs with min/max bounds; custom export calendar is skipped when native picker is available.
- Calendar day notes now auto-save consistently with other day fields.
- Privacy breadcrumb naming is aligned with app navigation (`Dashboard`/`Панель` for authenticated users).
- Profile save now supports inline HTMX success feedback, matching other settings forms.
- Mobile quick navigation tab bar was added for faster section switching.
- Day editor symptom chips now clear visual active state immediately when `Period day` is turned off (UI state now matches saved payload).
- Dashboard stale-cycle detection now prioritizes the owner-set cycle anchor date (`last_period_start`) to avoid showing stale data as factual.
- Language switch active pill styling was hardened for mobile, and frontend asset cache-busting versions were bumped to force fresh JS/CSS after deploy.
- Toast UX improved with longer visibility window and clearer close button affordance for manual dismissal.
- Russian copy polish: public text now consistently uses `надёжный` where applicable.
- Language switch active state now uses explicit `aria-current` styling and hard color values to avoid mobile active-pill label disappearance.
- HTMX save-status success banners are now dismissible (`×`) and no longer rely on `status-transient` 2s fade behavior.
- Stats page current-phase card now follows stale-cycle detection used on dashboard (shows `Unknown` with stale-phase hint when baseline is outdated).
- Save-status dismiss behavior is now guaranteed even without timing-sensitive HTMX hooks by rendering close controls server-side and adding `afterSettle` fallback handling in app JS.

## [0.1.0] - 2026-02-23

### Added
- Initial public release of Ovumcy.
- Privacy-first menstrual cycle tracking with:
  - daily logs (period day, flow, symptoms, notes),
  - cycle predictions (next period, ovulation, fertile window),
  - calendar and statistics views,
  - CSV/JSON export,
  - Russian/English localization.

[Unreleased]: https://github.com/terraincognita07/ovumcy/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/terraincognita07/ovumcy/releases/tag/v0.1.0
