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

func TestOnboardingPageRendersPersistedStep2Values(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "onboarding-values@example.com", "StrongPass1", false)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":     31,
		"period_length":    7,
		"auto_period_fill": true,
	}).Error; err != nil {
		t.Fatalf("update onboarding values: %v", err)
	}
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/onboarding", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("onboarding request failed: %v", err)
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
	if !strings.Contains(rendered, `<span x-text="cycleLength">31</span>`) {
		t.Fatalf("expected cycle label fallback text to include persisted value")
	}
	if !strings.Contains(rendered, `<span x-text="periodLength">7</span>`) {
		t.Fatalf("expected period label fallback text to include persisted value")
	}

	cycleInputPattern := regexp.MustCompile(`(?s)name="cycle_length".*?value="31"`)
	if !cycleInputPattern.MatchString(rendered) {
		t.Fatalf("expected cycle slider value attribute to be rendered from DB")
	}
	periodInputPattern := regexp.MustCompile(`(?s)name="period_length".*?value="7"`)
	if !periodInputPattern.MatchString(rendered) {
		t.Fatalf("expected period slider value attribute to be rendered from DB")
	}
	if !strings.Contains(rendered, `id="period-length"`) || !strings.Contains(rendered, `max="14"`) {
		t.Fatalf("expected onboarding period slider max=14")
	}
}
