package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
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
			Where(
				"user_id = ? AND date >= ? AND date < ?",
				updatedUser.ID,
				day.Format("2006-01-02"),
				day.AddDate(0, 0, 1).Format("2006-01-02"),
			).
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
		Where(
			"user_id = ? AND date >= ? AND date < ?",
			updatedUser.ID,
			today.Format("2006-01-02"),
			today.AddDate(0, 0, 1).Format("2006-01-02"),
		).
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
			Where(
				"user_id = ? AND date >= ? AND date < ?",
				user.ID,
				day.Format("2006-01-02"),
				day.AddDate(0, 0, 1).Format("2006-01-02"),
			).
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
		Where(
			"user_id = ? AND date >= ? AND date < ?",
			user.ID,
			afterEnd.Format("2006-01-02"),
			afterEnd.AddDate(0, 0, 1).Format("2006-01-02"),
		).
		First(&outside).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no onboarding log after finished period end date, got err=%v", err)
	}
}
