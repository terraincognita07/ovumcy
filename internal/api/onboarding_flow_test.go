package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/i18n"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestLoginRedirectsToOnboardingWhenOnboardingIncomplete(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "owner@example.com", "StrongPass1", false)

	form := url.Values{
		"email":    {user.Email},
		"password": {"StrongPass1"},
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
	if location := response.Header.Get("Location"); location != "/onboarding" {
		t.Fatalf("expected redirect to /onboarding, got %q", location)
	}
}

func TestOnboardingFlowCompletesWithOngoingPeriodRangeAndFlowNone(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "flow@example.com", "StrongPass1", false)

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	today := dateAtLocation(time.Now().In(time.UTC), time.UTC)
	stepDate := today.AddDate(0, 0, -2)
	stepDateRaw := stepDate.Format("2006-01-02")

	step1Form := url.Values{
		"last_period_start": {stepDateRaw},
		"period_status":     {onboardingPeriodStatusOngoing},
	}
	step1Request := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(step1Form.Encode()))
	step1Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	step1Request.Header.Set("HX-Request", "true")
	step1Request.Header.Set("Cookie", authCookie)

	step1Response, err := app.Test(step1Request, -1)
	if err != nil {
		t.Fatalf("step1 request failed: %v", err)
	}
	defer step1Response.Body.Close()
	if step1Response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected step1 status 204, got %d", step1Response.StatusCode)
	}

	step2Form := url.Values{
		"cycle_length":     {"30"},
		"period_length":    {"5"},
		"auto_period_fill": {"true"},
	}
	step2Request := httptest.NewRequest(http.MethodPost, "/onboarding/step2", strings.NewReader(step2Form.Encode()))
	step2Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	step2Request.Header.Set("HX-Request", "true")
	step2Request.Header.Set("Cookie", authCookie)

	step2Response, err := app.Test(step2Request, -1)
	if err != nil {
		t.Fatalf("step2 request failed: %v", err)
	}
	defer step2Response.Body.Close()
	if step2Response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected step2 status 204, got %d", step2Response.StatusCode)
	}

	step3Request := httptest.NewRequest(http.MethodPost, "/onboarding/complete", nil)
	step3Request.Header.Set("HX-Request", "true")
	step3Request.Header.Set("Cookie", authCookie)

	step3Response, err := app.Test(step3Request, -1)
	if err != nil {
		t.Fatalf("step3 request failed: %v", err)
	}
	defer step3Response.Body.Close()
	if step3Response.StatusCode != http.StatusOK {
		t.Fatalf("expected step3 status 200, got %d", step3Response.StatusCode)
	}
	if redirect := step3Response.Header.Get("HX-Redirect"); redirect != "/dashboard" {
		t.Fatalf("expected HX-Redirect /dashboard, got %q", redirect)
	}

	var updatedUser models.User
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if !updatedUser.OnboardingCompleted {
		t.Fatalf("expected onboarding to be completed")
	}
	if updatedUser.CycleLength != 30 {
		t.Fatalf("expected cycle length 30, got %d", updatedUser.CycleLength)
	}
	if updatedUser.PeriodLength != 5 {
		t.Fatalf("expected period length 5, got %d", updatedUser.PeriodLength)
	}
	if updatedUser.LastPeriodStart == nil {
		t.Fatalf("expected last period start to be saved")
	}
	if updatedUser.LastPeriodStart.Format("2006-01-02") != stepDateRaw {
		t.Fatalf("expected last period start %s, got %s", stepDateRaw, updatedUser.LastPeriodStart.Format("2006-01-02"))
	}
	if updatedUser.OnboardingPeriodStatus != "" {
		t.Fatalf("expected onboarding period status to be cleared after completion")
	}
	if updatedUser.OnboardingPeriodEnd != nil {
		t.Fatalf("expected onboarding period end to be cleared after completion")
	}

	for offset := 0; offset < 5; offset++ {
		day := stepDate.AddDate(0, 0, offset)
		var entry models.DailyLog
		if err := database.
			Where("user_id = ? AND substr(date, 1, 10) = ?", updatedUser.ID, day.Format("2006-01-02")).
			First(&entry).Error; err != nil {
			t.Fatalf("expected onboarding log for %s: %v", day.Format("2006-01-02"), err)
		}
		if !entry.IsPeriod {
			t.Fatalf("expected %s to be marked as period day", day.Format("2006-01-02"))
		}
		if entry.Flow != models.FlowNone {
			t.Fatalf("expected flow=none for %s, got %q", day.Format("2006-01-02"), entry.Flow)
		}
	}

	var todayLog models.DailyLog
	if err := database.
		Where("user_id = ? AND substr(date, 1, 10) = ?", updatedUser.ID, today.Format("2006-01-02")).
		First(&todayLog).Error; err != nil {
		t.Fatalf("expected today log to be included in onboarding range: %v", err)
	}
	if !todayLog.IsPeriod {
		t.Fatalf("expected today to be marked as period day when it is inside onboarding range")
	}
}

func TestOnboardingFlowFinishedPeriodUsesExactEndDateWithoutExtension(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "flow-finished@example.com", "StrongPass1", false)

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	startDay := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, -7)
	endDay := startDay.AddDate(0, 0, 3)

	step1Form := url.Values{
		"last_period_start": {startDay.Format("2006-01-02")},
		"period_status":     {onboardingPeriodStatusFinished},
		"period_end":        {endDay.Format("2006-01-02")},
	}
	step1Request := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(step1Form.Encode()))
	step1Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	step1Request.Header.Set("HX-Request", "true")
	step1Request.Header.Set("Cookie", authCookie)

	step1Response, err := app.Test(step1Request, -1)
	if err != nil {
		t.Fatalf("step1 request failed: %v", err)
	}
	defer step1Response.Body.Close()
	if step1Response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected step1 status 204, got %d", step1Response.StatusCode)
	}

	step2Form := url.Values{
		"cycle_length":     {"30"},
		"period_length":    {"8"},
		"auto_period_fill": {"true"},
	}
	step2Request := httptest.NewRequest(http.MethodPost, "/onboarding/step2", strings.NewReader(step2Form.Encode()))
	step2Request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	step2Request.Header.Set("HX-Request", "true")
	step2Request.Header.Set("Cookie", authCookie)

	step2Response, err := app.Test(step2Request, -1)
	if err != nil {
		t.Fatalf("step2 request failed: %v", err)
	}
	defer step2Response.Body.Close()
	if step2Response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected step2 status 204, got %d", step2Response.StatusCode)
	}

	step3Request := httptest.NewRequest(http.MethodPost, "/onboarding/complete", nil)
	step3Request.Header.Set("HX-Request", "true")
	step3Request.Header.Set("Cookie", authCookie)

	step3Response, err := app.Test(step3Request, -1)
	if err != nil {
		t.Fatalf("step3 request failed: %v", err)
	}
	defer step3Response.Body.Close()
	if step3Response.StatusCode != http.StatusOK {
		t.Fatalf("expected step3 status 200, got %d", step3Response.StatusCode)
	}

	for day := startDay; !day.After(endDay); day = day.AddDate(0, 0, 1) {
		var entry models.DailyLog
		if err := database.
			Where("user_id = ? AND substr(date, 1, 10) = ?", user.ID, day.Format("2006-01-02")).
			First(&entry).Error; err != nil {
			t.Fatalf("expected finished-period log for %s: %v", day.Format("2006-01-02"), err)
		}
		if !entry.IsPeriod {
			t.Fatalf("expected %s to be period day", day.Format("2006-01-02"))
		}
		if entry.Flow != models.FlowNone {
			t.Fatalf("expected flow=none for %s, got %q", day.Format("2006-01-02"), entry.Flow)
		}
	}

	afterEnd := endDay.AddDate(0, 0, 1)
	var outside models.DailyLog
	err = database.
		Where("user_id = ? AND substr(date, 1, 10) = ?", user.ID, afterEnd.Format("2006-01-02")).
		First(&outside).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no onboarding log after finished period end date, got err=%v", err)
	}
}

func TestOnboardingStep1RejectsFutureAndTooOldDates(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "step1-validation@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	futureDate := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, 1).Format("2006-01-02")
	futureForm := url.Values{
		"last_period_start": {futureDate},
		"period_status":     {onboardingPeriodStatusOngoing},
	}
	futureRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(futureForm.Encode()))
	futureRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	futureRequest.Header.Set("HX-Request", "true")
	futureRequest.Header.Set("Cookie", authCookie)

	futureResponse, err := app.Test(futureRequest, -1)
	if err != nil {
		t.Fatalf("future date request failed: %v", err)
	}
	defer futureResponse.Body.Close()
	if futureResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected future date status 400, got %d", futureResponse.StatusCode)
	}

	oldDate := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, -61).Format("2006-01-02")
	oldForm := url.Values{
		"last_period_start": {oldDate},
		"period_status":     {onboardingPeriodStatusOngoing},
	}
	oldRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(oldForm.Encode()))
	oldRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	oldRequest.Header.Set("HX-Request", "true")
	oldRequest.Header.Set("Cookie", authCookie)

	oldResponse, err := app.Test(oldRequest, -1)
	if err != nil {
		t.Fatalf("old date request failed: %v", err)
	}
	defer oldResponse.Body.Close()
	if oldResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected old date status 400, got %d", oldResponse.StatusCode)
	}
}

func TestOnboardingStep1RequiresPeriodEndForFinishedStatus(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "step1-finished-validation@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	stepDate := dateAtLocation(time.Now().In(time.UTC), time.UTC).AddDate(0, 0, -4).Format("2006-01-02")
	form := url.Values{
		"last_period_start": {stepDate},
		"period_status":     {onboardingPeriodStatusFinished},
		"period_end":        {""},
	}
	request := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("step1 finished-without-end request failed: %v", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}
}

func TestOnboardingStep2RejectsOutOfRangeValues(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "step2-validation@example.com", "StrongPass1", false)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	invalidCycleForm := url.Values{
		"cycle_length":  {"14"},
		"period_length": {"5"},
	}
	invalidCycleRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step2", strings.NewReader(invalidCycleForm.Encode()))
	invalidCycleRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	invalidCycleRequest.Header.Set("HX-Request", "true")
	invalidCycleRequest.Header.Set("Cookie", authCookie)

	invalidCycleResponse, err := app.Test(invalidCycleRequest, -1)
	if err != nil {
		t.Fatalf("invalid cycle request failed: %v", err)
	}
	defer invalidCycleResponse.Body.Close()
	if invalidCycleResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected invalid cycle status 400, got %d", invalidCycleResponse.StatusCode)
	}

	invalidPeriodForm := url.Values{
		"cycle_length":  {"29"},
		"period_length": {"11"},
	}
	invalidPeriodRequest := httptest.NewRequest(http.MethodPost, "/onboarding/step2", strings.NewReader(invalidPeriodForm.Encode()))
	invalidPeriodRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	invalidPeriodRequest.Header.Set("HX-Request", "true")
	invalidPeriodRequest.Header.Set("Cookie", authCookie)

	invalidPeriodResponse, err := app.Test(invalidPeriodRequest, -1)
	if err != nil {
		t.Fatalf("invalid period request failed: %v", err)
	}
	defer invalidPeriodResponse.Body.Close()
	if invalidPeriodResponse.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected invalid period status 400, got %d", invalidPeriodResponse.StatusCode)
	}
}

func newOnboardingTestApp(t *testing.T) (*fiber.App, *gorm.DB) {
	t.Helper()
	return newOnboardingTestAppWithCookieSecure(t, false)
}

func newOnboardingTestAppWithCookieSecure(t *testing.T, cookieSecure bool) (*fiber.App, *gorm.DB) {
	t.Helper()

	_, testFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve current test file path")
	}

	apiDir := filepath.Dir(testFile)
	internalDir := filepath.Dir(apiDir)
	templatesDir := filepath.Join(internalDir, "templates")
	localesDir := filepath.Join(internalDir, "i18n", "locales")
	databasePath := filepath.Join(t.TempDir(), "lume-onboarding-test.db")

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

	i18nManager, err := i18n.NewManager("en", localesDir)
	if err != nil {
		t.Fatalf("init i18n: %v", err)
	}

	handler, err := NewHandler(database, "test-secret-key", templatesDir, time.UTC, i18nManager, cookieSecure)
	if err != nil {
		t.Fatalf("init handler: %v", err)
	}

	app := fiber.New()
	app.Use(handler.LanguageMiddleware)
	RegisterRoutes(app, handler)
	return app, database
}

func createOnboardingTestUser(t *testing.T, database *gorm.DB, email string, password string, onboardingCompleted bool) models.User {
	t.Helper()

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	user := models.User{
		Email:               strings.ToLower(strings.TrimSpace(email)),
		PasswordHash:        string(passwordHash),
		Role:                models.RoleOwner,
		OnboardingCompleted: onboardingCompleted,
		CycleLength:         28,
		PeriodLength:        5,
		AutoPeriodFill:      true,
		CreatedAt:           time.Now().UTC(),
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func loginAndExtractAuthCookie(t *testing.T, app *fiber.App, email string, password string) string {
	t.Helper()

	form := url.Values{
		"email":    {email},
		"password": {password},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected login status 303, got %d", response.StatusCode)
	}

	for _, cookie := range response.Cookies() {
		if cookie.Name == "lume_auth" && cookie.Value != "" {
			return cookie.Name + "=" + cookie.Value
		}
	}

	t.Fatal("auth cookie is missing in login response")
	return ""
}
