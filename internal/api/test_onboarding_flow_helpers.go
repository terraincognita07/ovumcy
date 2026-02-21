package api

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

func submitOnboardingStep1(t *testing.T, app *fiber.App, authCookie string, form url.Values) {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/onboarding/step1", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("step1 request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected step1 status 204, got %d", response.StatusCode)
	}
}

func submitOnboardingStep2(t *testing.T, app *fiber.App, authCookie string, form url.Values) {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/onboarding/step2", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("step2 request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusNoContent {
		t.Fatalf("expected step2 status 204, got %d", response.StatusCode)
	}
}

func submitOnboardingComplete(t *testing.T, app *fiber.App, authCookie string) {
	t.Helper()

	request := httptest.NewRequest(http.MethodPost, "/onboarding/complete", nil)
	request.Header.Set("HX-Request", "true")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("step3 request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected step3 status 200, got %d", response.StatusCode)
	}
	if redirect := response.Header.Get("HX-Redirect"); redirect != "/dashboard" {
		t.Fatalf("expected HX-Redirect /dashboard, got %q", redirect)
	}
}

func assertOnboardingPeriodLogForDay(t *testing.T, database *gorm.DB, userID uint, day time.Time) {
	t.Helper()

	var entry models.DailyLog
	if err := database.
		Where(
			"user_id = ? AND date >= ? AND date < ?",
			userID,
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

func assertNoOnboardingLogForDay(t *testing.T, database *gorm.DB, userID uint, day time.Time) {
	t.Helper()

	var entry models.DailyLog
	err := database.
		Where(
			"user_id = ? AND date >= ? AND date < ?",
			userID,
			day.Format("2006-01-02"),
			day.AddDate(0, 0, 1).Format("2006-01-02"),
		).
		First(&entry).Error
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no onboarding log for %s, got err=%v", day.Format("2006-01-02"), err)
	}
}
