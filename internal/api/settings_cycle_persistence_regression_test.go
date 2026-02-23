package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestSettingsCycleUpdatePersistsAndRendersAfterReload(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-cycle-persist@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":     15,
		"period_length":    5,
		"auto_period_fill": false,
	}).Error; err != nil {
		t.Fatalf("set initial cycle values: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"cycle_length":      {"28"},
		"period_length":     {"6"},
		"auto_period_fill":  {"true"},
		"last_period_start": {"2026-02-10"},
	}
	updateRequest := httptest.NewRequest(http.MethodPost, "/settings/cycle", strings.NewReader(form.Encode()))
	updateRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	updateRequest.Header.Set("HX-Request", "true")
	updateRequest.Header.Set("Accept-Language", "en")
	updateRequest.Header.Set("Cookie", authCookie)

	updateResponse, err := app.Test(updateRequest, -1)
	if err != nil {
		t.Fatalf("settings cycle update request failed: %v", err)
	}
	defer updateResponse.Body.Close()

	if updateResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", updateResponse.StatusCode)
	}

	updateBody, err := io.ReadAll(updateResponse.Body)
	if err != nil {
		t.Fatalf("read update response: %v", err)
	}
	if !strings.Contains(string(updateBody), "status-ok") {
		t.Fatalf("expected htmx success status markup, got %q", string(updateBody))
	}

	persisted := models.User{}
	if err := database.Select("cycle_length", "period_length", "auto_period_fill", "last_period_start").First(&persisted, user.ID).Error; err != nil {
		t.Fatalf("load persisted user cycle values: %v", err)
	}
	if persisted.CycleLength != 28 {
		t.Fatalf("expected persisted cycle_length=28, got %d", persisted.CycleLength)
	}
	if persisted.PeriodLength != 6 {
		t.Fatalf("expected persisted period_length=6, got %d", persisted.PeriodLength)
	}
	if !persisted.AutoPeriodFill {
		t.Fatalf("expected persisted auto_period_fill=true")
	}
	if persisted.LastPeriodStart == nil || persisted.LastPeriodStart.Format("2006-01-02") != "2026-02-10" {
		t.Fatalf("expected persisted last_period_start=2026-02-10, got %v", persisted.LastPeriodStart)
	}

	settingsRequest := httptest.NewRequest(http.MethodGet, "/settings", nil)
	settingsRequest.Header.Set("Accept-Language", "en")
	settingsRequest.Header.Set("Cookie", authCookie)

	settingsResponse, err := app.Test(settingsRequest, -1)
	if err != nil {
		t.Fatalf("settings page request failed: %v", err)
	}
	defer settingsResponse.Body.Close()

	if settingsResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected settings status 200, got %d", settingsResponse.StatusCode)
	}

	settingsBody, err := io.ReadAll(settingsResponse.Body)
	if err != nil {
		t.Fatalf("read settings page body: %v", err)
	}
	rendered := string(settingsBody)
	if !strings.Contains(rendered, `x-data='settingsCycleForm({ cycleLength: 28, periodLength: 6, autoPeriodFill: true })'`) {
		t.Fatalf("expected settings cycle form state to include persisted values")
	}

	cycleInputPattern := regexp.MustCompile(`(?s)name="cycle_length".*?value="28"`)
	if !cycleInputPattern.MatchString(rendered) {
		t.Fatalf("expected cycle slider value 28 in rendered settings page")
	}
	periodInputPattern := regexp.MustCompile(`(?s)name="period_length".*?value="6"`)
	if !periodInputPattern.MatchString(rendered) {
		t.Fatalf("expected period slider value 6 in rendered settings page")
	}
	lastPeriodInputPattern := regexp.MustCompile(`(?s)name="last_period_start".*?value="2026-02-10"`)
	if !lastPeriodInputPattern.MatchString(rendered) {
		t.Fatalf("expected last_period_start date input to render persisted value")
	}
}
