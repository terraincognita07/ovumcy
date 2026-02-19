package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/i18n"
	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginInvalidCredentialsRedirectPreservesEmail(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "login-email@example.com", "StrongPass1", true)

	form := url.Values{
		"email":    {user.Email},
		"password": {"WrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}

	location := strings.TrimSpace(response.Header.Get("Location"))
	if location == "" {
		t.Fatalf("expected redirect location")
	}
	redirectURL, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if redirectURL.Path != "/login" {
		t.Fatalf("expected redirect path /login, got %q", redirectURL.Path)
	}
	if got := redirectURL.Query().Get("error"); got != "invalid credentials" {
		t.Fatalf("expected error query %q, got %q", "invalid credentials", got)
	}
	if got := redirectURL.Query().Get("email"); got != user.Email {
		t.Fatalf("expected email query %q, got %q", user.Email, got)
	}

	followRequest := httptest.NewRequest(http.MethodGet, location, nil)
	followRequest.Header.Set("Accept-Language", "en")
	followResponse, err := app.Test(followRequest, -1)
	if err != nil {
		t.Fatalf("follow-up login request failed: %v", err)
	}
	defer followResponse.Body.Close()

	if followResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected follow-up status 200, got %d", followResponse.StatusCode)
	}

	body, err := io.ReadAll(followResponse.Body)
	if err != nil {
		t.Fatalf("read follow-up body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `id="login-email"`) {
		t.Fatalf("expected login email input in page")
	}
	if !strings.Contains(rendered, `value="login-email@example.com"`) {
		t.Fatalf("expected login email input to keep previous value")
	}
}

func TestLanguageSwitchSetsCookieAndRendersMatchingHTMLLang(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	switchToEnglish := httptest.NewRequest(http.MethodGet, "/lang/en?next=/login", nil)
	switchToEnglish.Header.Set("Accept-Language", "ru")
	englishResponse, err := app.Test(switchToEnglish, -1)
	if err != nil {
		t.Fatalf("switch language request failed: %v", err)
	}
	defer englishResponse.Body.Close()

	if englishResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", englishResponse.StatusCode)
	}
	if location := englishResponse.Header.Get("Location"); location != "/login" {
		t.Fatalf("expected redirect to /login, got %q", location)
	}

	englishCookie := responseCookieValue(englishResponse.Cookies(), "lume_lang")
	if englishCookie != "en" {
		t.Fatalf("expected lume_lang cookie value %q, got %q", "en", englishCookie)
	}

	englishLoginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	englishLoginRequest.Header.Set("Cookie", "lume_lang="+englishCookie)
	englishLoginResponse, err := app.Test(englishLoginRequest, -1)
	if err != nil {
		t.Fatalf("english login request failed: %v", err)
	}
	defer englishLoginResponse.Body.Close()

	englishBody, err := io.ReadAll(englishLoginResponse.Body)
	if err != nil {
		t.Fatalf("read english login body: %v", err)
	}
	renderedEnglish := string(englishBody)
	if !strings.Contains(renderedEnglish, `<html lang="en"`) {
		t.Fatalf("expected login page html lang to be en")
	}
	if !strings.Contains(renderedEnglish, `applyHTMLLanguage(readCookie("lume_lang")`) {
		t.Fatalf("expected language sync script in base template")
	}
	if !strings.Contains(renderedEnglish, `data-required-message="Please fill out this field."`) {
		t.Fatalf("expected english required validation message in login form")
	}
	if !strings.Contains(renderedEnglish, `data-email-message="Please enter a valid email address."`) {
		t.Fatalf("expected english email validation message in login form")
	}

	switchToRussian := httptest.NewRequest(http.MethodGet, "/lang/ru?next=/login", nil)
	switchToRussian.Header.Set("Cookie", "lume_lang="+englishCookie)
	russianResponse, err := app.Test(switchToRussian, -1)
	if err != nil {
		t.Fatalf("switch back language request failed: %v", err)
	}
	defer russianResponse.Body.Close()

	russianCookie := responseCookieValue(russianResponse.Cookies(), "lume_lang")
	if russianCookie != "ru" {
		t.Fatalf("expected lume_lang cookie value %q, got %q", "ru", russianCookie)
	}

	russianLoginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	russianLoginRequest.Header.Set("Cookie", "lume_lang="+russianCookie)
	russianLoginResponse, err := app.Test(russianLoginRequest, -1)
	if err != nil {
		t.Fatalf("russian login request failed: %v", err)
	}
	defer russianLoginResponse.Body.Close()

	russianBody, err := io.ReadAll(russianLoginResponse.Body)
	if err != nil {
		t.Fatalf("read russian login body: %v", err)
	}
	if !strings.Contains(string(russianBody), `<html lang="ru"`) {
		t.Fatalf("expected login page html lang to be ru")
	}
	if !strings.Contains(string(russianBody), `data-required-message="Заполните это поле."`) {
		t.Fatalf("expected russian required validation message in login form")
	}
	if !strings.Contains(string(russianBody), `data-email-message="Введите корректный email адрес."`) {
		t.Fatalf("expected russian email validation message in login form")
	}
}

func TestDashboardSymptomsNotesPanelUsesSavedSymptomsAndNotesState(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "dashboard-journal@example.com", "StrongPass1", true)

	symptoms := []models.SymptomType{
		{UserID: user.ID, Name: "Custom cramps", Icon: "A", Color: "#FF7755"},
		{UserID: user.ID, Name: "Custom headache", Icon: "B", Color: "#55AAFF"},
	}
	if err := database.Create(&symptoms).Error; err != nil {
		t.Fatalf("create symptoms: %v", err)
	}

	today := dateAtLocation(time.Now().In(time.UTC), time.UTC)
	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       today,
		IsPeriod:   false,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{symptoms[0].ID, symptoms[1].ID},
		Notes:      "Remember to hydrate",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)
	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, "Symptoms and notes") {
		t.Fatalf("expected symptoms and notes panel header")
	}
	if !strings.Contains(rendered, `x-for="(symptom, index) in activeSymptoms"`) {
		t.Fatalf("expected dynamic symptoms list renderer in panel")
	}
	if !strings.Contains(rendered, `data-symptom-label="Custom cramps"`) {
		t.Fatalf("expected symptom checkbox metadata for dashboard preview")
	}
	if !strings.Contains(rendered, `data-symptom-label="Custom headache"`) {
		t.Fatalf("expected second symptom checkbox metadata for dashboard preview")
	}
	if !strings.Contains(rendered, `x-model="notesPreview"`) {
		t.Fatalf("expected notes field binding for dashboard preview")
	}
	if !strings.Contains(rendered, "Remember to hydrate") {
		t.Fatalf("expected saved note to be rendered on dashboard")
	}
}

func TestCalendarDayPanelFlowControlsDependOnPeriodToggle(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-flow-toggle@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	panelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	panelRequest.Header.Set("Accept-Language", "en")
	panelRequest.Header.Set("Cookie", authCookie)

	panelResponse, err := app.Test(panelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer panelResponse.Body.Close()

	if panelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", panelResponse.StatusCode)
	}

	body, err := io.ReadAll(panelResponse.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	rendered := string(body)

	if !strings.Contains(rendered, `x-data="{ isPeriod: false }"`) {
		t.Fatalf("expected calendar panel form to initialize period state")
	}
	if !strings.Contains(rendered, `x-model="isPeriod"`) {
		t.Fatalf("expected period toggle to drive alpine state")
	}
	if !strings.Contains(rendered, `x-cloak x-show="isPeriod" :disabled="!isPeriod"`) {
		t.Fatalf("expected flow fieldset to be shown/enabled only when period is selected")
	}
	if strings.Count(rendered, `:disabled="!isPeriod"`) != 1 {
		t.Fatalf("expected only flow controls to depend on period toggle")
	}
	if !strings.Contains(rendered, `name="symptom_ids"`) {
		t.Fatalf("expected symptoms controls to stay available regardless of period toggle")
	}
}

func TestBuildCalendarDaysUsesLatestLogPerDateDeterministically(t *testing.T) {
	handler := &Handler{location: time.UTC}
	monthStart := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)
	now := time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC)

	logs := []models.DailyLog{
		{
			ID:       20,
			Date:     time.Date(2026, time.February, 17, 20, 0, 0, 0, time.UTC),
			IsPeriod: false,
			Flow:     models.FlowNone,
		},
		{
			ID:       10,
			Date:     time.Date(2026, time.February, 17, 8, 0, 0, 0, time.UTC),
			IsPeriod: true,
			Flow:     models.FlowMedium,
		},
		{
			ID:       30,
			Date:     time.Date(2026, time.February, 18, 9, 0, 0, 0, time.UTC),
			IsPeriod: true,
			Flow:     models.FlowMedium,
		},
		{
			ID:       31,
			Date:     time.Date(2026, time.February, 18, 9, 0, 0, 0, time.UTC),
			IsPeriod: false,
			Flow:     models.FlowNone,
		},
	}

	days := handler.buildCalendarDays(monthStart, logs, services.CycleStats{}, now)

	day17 := findCalendarDayByDateString(t, days, "2026-02-17")
	if day17.IsPeriod {
		t.Fatalf("expected 2026-02-17 period=false from latest log, got true")
	}

	day18 := findCalendarDayByDateString(t, days, "2026-02-18")
	if day18.IsPeriod {
		t.Fatalf("expected 2026-02-18 period=false from highest id tie-breaker, got true")
	}
}

func TestFetchLogByDateFindsZuluStoredRowForLocalCalendarDay(t *testing.T) {
	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}

	apiDir := filepath.Dir(testFile)
	internalDir := filepath.Dir(apiDir)
	templatesDir := filepath.Join(internalDir, "templates")
	localesDir := filepath.Join(internalDir, "i18n", "locales")
	databasePath := filepath.Join(t.TempDir(), "lume-zulu-fetch.db")

	database, err := db.OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := models.User{
		Email:               "zulu-fetch@example.com",
		PasswordHash:        string(passwordHash),
		Role:                models.RoleOwner,
		OnboardingCompleted: true,
		CycleLength:         28,
		PeriodLength:        5,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	now := time.Now().UTC()
	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		"2026-02-17T00:00:00Z",
		true,
		models.FlowLight,
		"[]",
		"",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert zulu row: %v", err)
	}

	i18nManager, err := i18n.NewManager("en", localesDir)
	if err != nil {
		t.Fatalf("init i18n: %v", err)
	}

	moscow := time.FixedZone("UTC+3", 3*60*60)
	handler, err := NewHandler(database, "test-secret-key", templatesDir, moscow, i18nManager)
	if err != nil {
		t.Fatalf("init handler: %v", err)
	}

	day, err := parseDayParam("2026-02-17", moscow)
	if err != nil {
		t.Fatalf("parse day: %v", err)
	}

	entry, err := handler.fetchLogByDate(user.ID, day)
	if err != nil {
		t.Fatalf("fetchLogByDate: %v", err)
	}

	if !entry.IsPeriod {
		t.Fatalf("expected is_period=true for local day 2026-02-17")
	}
	if entry.Flow != models.FlowLight {
		t.Fatalf("expected flow %q, got %q", models.FlowLight, entry.Flow)
	}
}

func responseCookieValue(cookies []*http.Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}
	return ""
}

func findCalendarDayByDateString(t *testing.T, days []CalendarDay, date string) CalendarDay {
	t.Helper()
	for _, day := range days {
		if day.DateString == date {
			return day
		}
	}
	t.Fatalf("calendar day %s not found", date)
	return CalendarDay{}
}
