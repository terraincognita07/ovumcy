---
name: "release-plan"
description: "Generate an end-to-end release regression plan for ovumcy, including backend tests, frontend checks, privacy/security-sensitive areas, and manual UI scenarios."
---

# Skill: release-plan

## Purpose

Help the user prepare a structured, repeatable release and regression plan for the ovumcy project:
- enumerate what must be tested before a release,
- separate automated checks from manual UI flows,
- highlight privacy/security-sensitive areas and high-risk changes.

## Inputs

- A short description of the release scope:
  - features changed,
  - areas touched (backend / frontend / auth / exports / settings / calendar / roles),
  - known risks or migration notes (if any).

## Workflow

1. Clarify the release scope
   - Ask the user for:
     - a short summary of what changed in this release,
     - which modules or features were touched,
     - whether there are DB migrations, auth/permission changes, or export-related changes.

2. Identify impacted domains
   - Map the described changes into ovumcy domains, for example:
     - Authentication & sessions
     - Roles and access control (owner vs partner)
     - Core tracking flows (logging data, editing, deleting)
     - Calendar and stats
     - Settings (cycle, profile, privacy)
     - Data export / import
     - UI and i18n (ru/en)
   - Explicitly mark which domains are **privacy/security-critical**.

3. Propose automated checks
   - For backend:
     - propose running `go test ./...` or more targeted packages if the scope is narrow.
   - For frontend:
     - propose running `npm run lint` and `npm run build`.
   - If the project defines additional scripts (lint, type-check, etc.), include them once discovered.
   - Ask the user which commands are acceptable in this environment before suggesting them as part of the plan.

4. Define manual UI regression flows
   - For each impacted domain, generate concrete manual flows, separated by role and device type:
     - Owner (desktop): main happy paths + edge cases.
     - Owner (mobile viewport): critical flows (auth, dashboard, calendar, settings, export).
     - Partner/viewer (desktop + mobile): ensure read-only invariants and proper privacy sanitization.
   - Use clear, step-by-step scenarios, e.g.:
     - “Owner: sign up → complete onboarding → log a few days → open calendar → verify stats.”
     - “Partner: open shared link → navigate between views → ensure no private fields are visible.”

5. Highlight privacy and security checks
   - For any flow that touches health-related or sensitive data, add explicit checks:
     - no leakage of private fields in UI,
     - correct behavior when access is denied,
     - correct handling of language/locale.
   - Call out any areas where there is currently no automated coverage and suggest adding tests in the future.

6. Produce the final release plan
   - Output the plan as structured sections:
     - Summary of scope and risk areas
     - Automated checks (commands and expected outcomes)
     - Manual UI flows (by role and device type)
     - Privacy/security-specific checks
     - Known limitations / TODOs for future automation
   - Make the plan concise enough to be followed in one sitting, but explicit enough to be repeatable.

## Constraints

- Do not assume you can run commands without explicit user approval; always present them as suggestions first.
- Do not promise full coverage if the codebase does not have relevant automated tests; clearly state gaps and propose follow-up work instead.
- Always respect global and project AGENTS instructions, especially around privacy, security, and architectural boundaries.
