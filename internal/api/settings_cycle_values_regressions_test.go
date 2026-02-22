package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestSettingsPageRendersPersistedCycleValues(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-values@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":     29,
		"period_length":    6,
		"auto_period_fill": true,
	}).Error; err != nil {
		t.Fatalf("update cycle values: %v", err)
	}
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `x-data='typeof settingsCycleForm === "function" ? settingsCycleForm({ cycleLength: 29, periodLength: 6, autoPeriodFill: true }) : { cycleLength: 29, periodLength: 6, autoPeriodFill: true }'`) {
		t.Fatalf("expected settings cycle form state to include persisted values")
	}
	if !strings.Contains(rendered, `<span x-text="cycleLength">29</span>`) {
		t.Fatalf("expected cycle label fallback text to include persisted value")
	}
	if !strings.Contains(rendered, `<span x-text="periodLength">6</span>`) {
		t.Fatalf("expected period label fallback text to include persisted value")
	}
	if !strings.Contains(rendered, `x-show="(cycleLength - periodLength) < 8"`) {
		t.Fatalf("expected settings cycle form to include hard-validation state for incompatible cycle values")
	}
	if !strings.Contains(rendered, `'btn--disabled': (cycleLength - periodLength) < 8`) {
		t.Fatalf("expected settings save button to include disabled visual state class binding")
	}

	cycleInputPattern := regexp.MustCompile(`(?s)name="cycle_length".*?value="29"`)
	if !cycleInputPattern.MatchString(rendered) {
		t.Fatalf("expected cycle slider value attribute to be rendered from DB")
	}
	periodInputPattern := regexp.MustCompile(`(?s)name="period_length".*?value="6"`)
	if !periodInputPattern.MatchString(rendered) {
		t.Fatalf("expected period slider value attribute to be rendered from DB")
	}
	if !strings.Contains(rendered, `id="settings-period-length"`) || !strings.Contains(rendered, `max="14"`) {
		t.Fatalf("expected settings period slider max=14")
	}
}
