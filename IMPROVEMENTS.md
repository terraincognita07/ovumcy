# Lume - UI/UX Improvements

## Changes to implement

### 1. Smart login/register detection

**Current problem:**
- Two forms side-by-side (login + register) on every visit
- Confusing for first-time users

**Solution:**
- Backend: add GET /api/auth/setup-status endpoint that returns {needs_setup: bool}
- Frontend: on page load, check setup-status:
  - If needs_setup=true → show ONLY registration form ("Create your account")
  - If needs_setup=false → show ONLY login form ("Welcome back")
- No manual switching, automatic detection

### 2. Internationalization (i18n)

**Languages:** Russian (default) + English

**Implementation:**
- Create internal/i18n/ package with translation files:
  - ru.json
  - en.json
- Add language switcher in navbar (RU/EN toggle)
- Store language preference in cookie
- Translate ALL UI strings:
  - Navigation (Dashboard, Calendar, Stats)
  - Form labels (Email, Password, Period day, Flow, Symptoms, Notes)
  - Buttons (Save, Login, Logout, Create account)
  - Phase names (menstrual, follicular, ovulation, luteal)
  - Stats labels (Average cycle, Median cycle, etc.)

**Translation keys structure:**
json
{
  "nav.dashboard": "Панель",
  "nav.calendar": "Календарь",
  "nav.stats": "Статистика",
  "auth.login": "Вход",
  "auth.register": "Создать аккаунт",
  "auth.email": "Email",
  "auth.password": "Пароль",
  "dashboard.current_phase": "Текущая фаза",
  "dashboard.cycle_day": "День цикла",
  "dashboard.next_period": "Следующие месячные",
  "dashboard.ovulation": "Овуляция",
  "dashboard.period_day": "День месячных",
  "dashboard.flow": "Интенсивность",
  "dashboard.flow.none": "Нет",
  "dashboard.flow.light": "Слабая",
  "dashboard.flow.medium": "Средняя",
  "dashboard.flow.heavy": "Сильная",
  "dashboard.symptoms": "Симптомы",
  "dashboard.notes": "Заметки",
  "dashboard.save_today": "Сохранить",
  "symptoms.acne": "Акне",
  "symptoms.bloating": "Вздутие",
  "symptoms.breast_tenderness": "Болезненность груди",
  "symptoms.cramps": "Спазмы",
  "symptoms.fatigue": "Усталость",
  "symptoms.headache": "Головная боль",
  "symptoms.mood_swings": "Перепады настроения",
  "phases.unknown": "Неизвестно",
  "phases.menstrual": "Менструальная",
  "phases.follicular": "Фолликулярная",
  "phases.ovulation": "Овуляция",
  "phases.luteal": "Лютеиновая"
}


Pass current language to templates via context.

### 3. Warm "journal-like" design system

**Goal:** Cozy, personal, handwritten-diary aesthetic. NOT clinical, NOT tech-minimalist.

**Color palette:**
- Background: #FFF9F0 (warm cream)
- Cards: #FFFFFF with soft shadows
- Primary accent: #D4A574 (warm caramel/brown)
- Secondary accent: #E8C4A8 (light peach)
- Period red: #C7756D (muted terracotta, not bright red)
- Ovulation yellow: #F4D58D (soft butter yellow)
- Fertility window: #B8D4C1 (sage green)
- Text: #5A4A3A (warm dark brown, not black)
- Muted text: #9B8B7A

**Typography:**
- Headings: slightly rounded sans-serif (or Google Font "Quicksand" / "Comfortaa" bundled locally)
- Body: clean readable sans (Inter or similar)
- Sizes: generous, comfortable (16px base)

**Components style:**
- Border radius: 12px–16px (soft, rounded)
- Shadows: soft, warm-toned (not gray shadows)
- Buttons: pill-shaped, with subtle texture/gradient
- Inputs: soft borders, focus glow (not sharp blue outline)
- Checkboxes: custom styled with rounded corners and warm colors
- Radio buttons: large, tactile, with icons inside

**Layout:**
- More whitespace (breathing room)
- Cards with gentle elevation
- Section dividers: subtle, decorative (not harsh lines)

**Iconography:**
- Use emoji for symptoms (keep current 🩸🎈💔🤕😴😢)
- Phase icons: add decorative touches (moon phases, flowers, sun)

**Animation:**
- Smooth transitions (200–300ms ease-out)
- Hover: gentle lift + shadow increase
- No jarring animations

**Sample CSS updates:**

css
/* Warm palette */
:root {
  --bg-primary: #FFF9F0;
  --bg-card: #FFFFFF;
  --text-primary: #5A4A3A;
  --text-muted: #9B8B7A;
  --accent-primary: #D4A574;
  --accent-secondary: #E8C4A8;
  --period-color: #C7756D;
  --ovulation-color: #F4D58D;
  --fertile-color: #B8D4C1;
  --shadow: 0 2px 8px rgba(212, 165, 116, 0.15);
}

body {
  background: var(--bg-primary);
  color: var(--text-primary);
  font-family: 'Inter', -apple-system, sans-serif;
  font-size: 16px;
}

.card {
  background: var(--bg-card);
  border-radius: 16px;
  box-shadow: var(--shadow);
  padding: 1.5rem;
  transition: transform 200ms ease-out, box-shadow 200ms ease-out;
}

.card:hover {
  transform: translateY(-2px);
  box-shadow: 0 4px 16px rgba(212, 165, 116, 0.25);
}

.btn-primary {
  background: linear-gradient(135deg, var(--accent-primary), var(--accent-secondary));
  color: white;
  border: none;
  border-radius: 24px;
  padding: 0.75rem 2rem;
  font-weight: 500;
  cursor: pointer;
  transition: all 200ms ease-out;
}

.btn-primary:hover {
  transform: translateY(-1px);
  box-shadow: 0 4px 12px rgba(212, 165, 116, 0.3);
}

input[type="text"],
input[type="password"],
input[type="email"],
textarea {
  background: var(--bg-card);
  border: 2px solid var(--accent-secondary);
  border-radius: 12px;
  padding: 0.75rem 1rem;
  color: var(--text-primary);
  font-size: 16px;
  transition: border-color 200ms ease-out;
}

input:focus,
textarea:focus {
  outline: none;
  border-color: var(--accent-primary);
  box-shadow: 0 0 0 3px rgba(212, 165, 116, 0.1);
}


**Additional touches:**
- Add subtle paper-like texture to background (optional CSS pattern)
- Period day checkbox: custom styled as a rounded toggle with period icon
- Calendar: softer grid lines, rounded day cells
- Stats charts: use warm color palette, rounded bars

### 4. File changes needed

**Backend:**
- internal/api/handlers.go: add SetupStatusHandler
- internal/i18n/i18n.go: i18n loader and helper functions
- internal/i18n/locales/ru.json
- internal/i18n/locales/en.json
- internal/api/middleware.go: add LanguageMiddleware (read cookie, set context)

**Frontend:**
- web/static/css/input.css: rewrite with warm palette
- internal/templates/base.html: add language switcher, update colors
- internal/templates/login.html: conditional render (register OR login, not both)
- internal/templates/dashboard.html: translate all strings using template functions
- internal/templates/calendar.html: translate, apply warm colors
- internal/templates/stats.html: translate, warm chart colors

**Config:**
- Add DEFAULT_LANGUAGE env var (default: "ru")

### 5. Implementation priority

1. Setup-status detection (simplest UX win)
2. Warm design system (CSS rewrite)
3. i18n backend + frontend wiring
4. Translation files (ru.json, en.json)
5. Language switcher UI

---

## Expected result

- First visit: warm, inviting registration form in Russian (or browser language)
- Returning visit: clean login form, no clutter
- UI feels personal, like a handwritten journal
- Language toggle in navbar (RU ↔ EN)
- All text translatable
- Colors are soft, warm, cozy
- No resemblance to commercial period trackers

---

END OF IMPROVEMENTS SPEC
