package api

import (
	"io"
	"net/http"
	"net/http/httptest"
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

func TestCalendarDayPanelDeleteEntryUsesConfirmForm(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-confirm@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	logEntry := models.DailyLog{
		UserID:   user.ID,
		Date:     time.Date(2026, time.February, 17, 0, 0, 0, 0, time.UTC),
		IsPeriod: true,
		Flow:     models.FlowMedium,
		Notes:    "entry",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `hx-delete="/api/log/delete?date=2026-02-17&source=calendar"`) {
		t.Fatalf("expected delete endpoint in day panel")
	}
	if !strings.Contains(rendered, `data-confirm="Are you sure you want to delete this entry?"`) {
		t.Fatalf("expected confirm prompt on calendar delete entry action")
	}
	if !strings.Contains(rendered, `data-confirm-accept="Yes, delete"`) {
		t.Fatalf("expected confirm accept label on calendar delete entry action")
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

	if !strings.Contains(rendered, `x-data='dayEditorForm({ isPeriod: false })'`) {
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

func TestCalendarDayPanelUsesLanguageSpecificSymptomLabelClass(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-symptom-label-class@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Breast tenderness",
		Icon:   "💗",
		Color:  "#D98395",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	enRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	enRequest.Header.Set("Accept-Language", "en")
	enRequest.Header.Set("Cookie", authCookie)
	enResponse, err := app.Test(enRequest, -1)
	if err != nil {
		t.Fatalf("english panel request failed: %v", err)
	}
	defer enResponse.Body.Close()

	enBody, err := io.ReadAll(enResponse.Body)
	if err != nil {
		t.Fatalf("read english panel body: %v", err)
	}
	if !strings.Contains(string(enBody), `symptom-label symptom-label-nowrap`) {
		t.Fatalf("expected nowrap symptom class for english labels")
	}

	ruRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	ruRequest.Header.Set("Accept-Language", "ru")
	ruRequest.Header.Set("Cookie", authCookie)
	ruResponse, err := app.Test(ruRequest, -1)
	if err != nil {
		t.Fatalf("russian panel request failed: %v", err)
	}
	defer ruResponse.Body.Close()

	ruBody, err := io.ReadAll(ruResponse.Body)
	if err != nil {
		t.Fatalf("read russian panel body: %v", err)
	}
	if strings.Contains(string(ruBody), `symptom-label symptom-label-nowrap`) {
		t.Fatalf("did not expect nowrap class for russian labels")
	}
	if !strings.Contains(string(ruBody), `class="symptom-label">`) {
		t.Fatalf("expected default symptom label class for russian locale")
	}
}

func TestCalendarPageKeepsSelectedDayFromQueryAndBootstrapsEditor(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-selected-day-query@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/calendar?month=2026-02&day=2026-02-17", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("calendar request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read calendar body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `selectedDate: "2026-02-17"`) {
		t.Fatalf("expected selected day in alpine state from day query")
	}
	if !strings.Contains(rendered, `hx-get="/calendar/day/2026-02-17"`) || !strings.Contains(rendered, `hx-trigger="load"`) {
		t.Fatalf("expected day editor bootstrap request for selected day")
	}
	if !strings.Contains(rendered, `next=%2Fcalendar%3Fmonth%3D2026-02%26day%3D2026-02-17`) {
		t.Fatalf("expected language switch links to preserve selected day in next param")
	}
	if !strings.Contains(rendered, `<script defer src="/static/js/app.js"></script>`) {
		t.Fatalf("expected shared app script to keep language links in sync")
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
	handler, err := NewHandler(database, "test-secret-key", templatesDir, moscow, i18nManager, false)
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
