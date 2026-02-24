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
