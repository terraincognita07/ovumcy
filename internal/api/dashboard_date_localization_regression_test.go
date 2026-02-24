package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestDashboardEnglishRendersLocalizedPredictionDates(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "dashboard-date-localization@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	lastPeriodStart := dateAtLocation(time.Now().UTC(), time.UTC).AddDate(0, 0, -8)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":      28,
		"period_length":     5,
		"last_period_start": lastPeriodStart,
	}).Error; err != nil {
		t.Fatalf("update user cycle context: %v", err)
	}

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

	nextPeriodPattern := regexp.MustCompile(`(?s)Next period</p>\s*<p class="stat-value mt-3">[A-Z][a-z]{2} \d{1,2}, \d{4}</p>`)
	if !nextPeriodPattern.MatchString(rendered) {
		t.Fatalf("expected English-localized next period date format in dashboard card")
	}

	ovulationPattern := regexp.MustCompile(`(?s)Ovulation</p>\s*<p class="stat-value mt-3">[A-Z][a-z]{2} \d{1,2}, \d{4}</p>`)
	if !ovulationPattern.MatchString(rendered) {
		t.Fatalf("expected English-localized ovulation date format in dashboard card")
	}
}
