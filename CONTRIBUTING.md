# Contributing to Ovumcy

Thanks for contributing.

## Development Setup

1. Install Go and Node.js.
2. Install frontend deps:

```bash
npm ci
```

3. Run checks locally:

```bash
go test ./...
npm run lint:js
npm run build
```

4. Start app locally:

```bash
go run ./cmd/ovumcy
```

## Pull Request Rules

- Keep changes scoped and atomic.
- Add/adjust tests for behavioral changes.
- Keep `internal/i18n/locales/en.json` and `internal/i18n/locales/ru.json` in sync.
- Do not introduce legacy compatibility paths unless explicitly required.

## Commit Style

Use imperative commit messages, e.g.:

- `Fix calendar ovulation tag precedence`
- `Pin staticcheck version in CI`
