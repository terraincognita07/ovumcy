# Lume - Self-Hosted Menstrual Cycle Tracker

You are an expert full-stack engineer.  
Generate a complete, production-ready MVP of a **self-hosted menstrual cycle tracker** as a Dockerized web app.

## High-level product goals

- Private, self-hosted period / cycle tracker, similar in functionality to Flo/Clue but much simpler.
- Single-tenant app for one primary user (the woman) plus an optional partner read-only account.
- All data stored locally (SQLite database) with NO external analytics, NO third-party SDKs, and NO calls to external APIs (except optional Telegram for notifications).
- Deployed via docker-compose up, exposing a single HTTP port (e.g. 8080).

---

## Tech stack

- Backend:
  - Go 1.26+
  - Fiber v2
  - GORM (ORM)
  - SQLite (file on disk)
  - JWT cookie-based auth

- Frontend:
  - HTML templates (html/template)
  - HTMX for dynamic updates
  - TailwindCSS for styling (bundled locally, no CDN)
  - Alpine.js for minimal interactivity

- Containerization:
  - Multi-stage Dockerfile
  - docker-compose.yml with volume for SQLite DB

---

## Repository structure

Use existing structure in cmd/lume/ and internal/:

- cmd/lume/main.go – entry point
- internal/api/handlers.go – HTTP handlers
- internal/api/routes.go – routing
- internal/api/middleware.go – auth middleware
- internal/models/user.go – User model
- internal/models/symptom.go – SymptomType model
- internal/models/daily_log.go – DailyLog model
- internal/db/sqlite.go – DB connection
- internal/services/cycles.go – cycle logic & predictions
- internal/services/notifications.go – Telegram notifications (optional)
- internal/templates/ – HTML templates
  - base.html – layout
  - login.html
  - dashboard.html
  - calendar.html
  - stats.html
- web/static/css/ – TailwindCSS compiled
- web/static/js/ – HTMX, Alpine.js
- migrations/001_init.sql – SQL schema
- docker/Dockerfile
- docker/docker-compose.yml

---

## Data model

Implement GORM models:

### User
- ID uint (PK)
- Email string (unique, not null)
- PasswordHash string
- Role string (owner / partner)
- CreatedAt time.Time

### SymptomType
- ID uint (PK)
- UserID uint (FK)
- Name string
- Icon string (emoji)
- Color string (hex)
- IsBuiltin bool

### DailyLog
- ID uint (PK)
- UserID uint (FK)
- Date time.Time (unique per user)
- IsPeriod bool
- Flow string (none / light / medium / heavy)
- SymptomIDs []uint (JSON)
- Notes string

---

## Cycle & prediction logic

Implement in internal/services/cycles.go:

1. Detect cycles from DailyLog:
   - Cycle starts on first is_period=true after 5+ non-period days
   - Cycle ends day before next cycle start

2. Compute statistics:
   - Last N=6 cycle lengths
   - Average and median cycle length
   - Average period length

3. Predictions:
   - next_period_start = last_period_start + median_cycle_length
   - luteal_phase_days = 14 (configurable)
   - ovulation_date = next_period_start - luteal_phase_days
   - fertility_window = ovulation_date - 5 to ovulation_date + 1

Return as struct:
go
type CycleStats struct {
    CurrentCycleDay      int
    CurrentPhase         string
    AverageCycleLength   float64
    MedianCycleLength    int
    AveragePeriodLength  float64
    LastPeriodStart      time.Time
    NextPeriodStart      time.Time
    OvulationDate        time.Time
    FertilityWindowStart time.Time
    FertilityWindowEnd   time.Time
}


---

## Backend API

### Auth (/api/auth)

POST /api/auth/register
- Create first owner user (only if no users exist)
- Input: email, password
- Output: set JWT cookie

POST /api/auth/login
- Input: email, password
- Output: set JWT cookie

POST /api/auth/logout
- Clear cookie

### Daily logs (/api/days)

GET /api/days?from=YYYY-MM-DD&to=YYYY-MM-DD
- Returns []DailyLog for date range

GET /api/days/:date
- Returns DailyLog for specific date

POST /api/days/:date
- Create/update log
- Body: {is_period, flow, symptom_ids, notes}

### Symptoms (/api/symptoms)

GET /api/symptoms
- List user's symptom types

POST /api/symptoms
- Create custom symptom

DELETE /api/symptoms/:id
- Delete custom symptom

### Stats (/api/stats/overview)

GET /api/stats/overview
- Returns CycleStats JSON

---

## Frontend requirements

### HTML Templates with HTMX

Use html/template with layout inheritance.

### Pages:

1. /login
   - Simple form
   - POST to /api/auth/login
   - Redirect to / on success

2. / (Dashboard)
   - Show CycleStats summary
   - Today editor form:
     - Period toggle
     - Flow radio buttons
     - Symptom checkboxes
     - Notes textarea (hidden for partner role)
   - HTMX: POST to /api/days/{today} on save

3. /calendar
   - Month view calendar grid
   - Color-coded days:
     - Red: actual period
     - Pink: predicted period
     - Yellow: fertility window
     - Icon on ovulation date
   - Click day: HTMX load day editor in side panel
   - Partner role: read-only view

4. /stats
   - Display CycleStats
   - Line chart of cycle lengths (Chart.js or similar, bundled)
   - Symptom frequency list

### Styling

- TailwindCSS compiled and bundled
- Dark theme default
- Responsive (mobile-friendly)

---

## Telegram notifications

In internal/services/notifications.go:

- Read env: TELEGRAM_BOT_TOKEN, TELEGRAM_CHAT_ID
- If set, send reminders:
  - N days before predicted period
  - Fertility window start (optional)
- Use cron job or background goroutine
- If env vars missing, do nothing

---

## Docker

### Dockerfile (multi-stage)

Stage 1: Build TailwindCSS
- Use node:20-alpine
- Build web/static/css/

Stage 2: Build Go app
- Use golang:1.26-alpine as builder
- Build cmd/lume/main.go
- Copy templates and static files

Stage 3: Runtime
- Use alpine:latest
- Copy binary, templates, static
- Expose 8080
- CMD: /app/lume

### docker-compose.yml

yaml
version: "3.9"
services:
  lume:
    build:
      context: .
      dockerfile: docker/Dockerfile
    container_name: lume
    environment:
      - TZ=Europe/Belgrade
      - SECRET_KEY=change_me_in_production
      - TELEGRAM_BOT_TOKEN=
      - TELEGRAM_CHAT_ID=
    volumes:
      - ./data:/app/data
    ports:
      - "8080:8080"
    restart: unless-stopped


SQLite DB at /app/data/lume.db

---

## Requirements

- No external network calls (except Telegram if configured)
- No external CDNs
- Handle timezone via env TZ
- Clean code with Go best practices
- Type safety with GORM

## Implementation order

1. Database models and migrations
2. Auth system with JWT
3. Daily log CRUD
4. Cycle prediction service
5. HTML templates with HTMX
6. Calendar UI
7. Stats page
8. Telegram notifications
9. Docker build
10. Testing

---

## Built-in symptoms to seed

Create these SymptomType records on first run:
- Cramps (🩸, #FF4444)
- Headache (🤕, #FFA500)
- Mood swings (😢, #9B59B6)
- Bloating (🎈, #3498DB)
- Fatigue (😴, #95A5A6)
- Breast tenderness (💔, #E91E63)
- Acne (🔴, #E74C3C)

---

## Security

- Password hashing with bcrypt
- JWT tokens with expiry
- CSRF protection
- SQL injection prevention (GORM parameterized queries)
- Input validation
- Partner role: hide symptoms and notes

---

END OF SPEC
