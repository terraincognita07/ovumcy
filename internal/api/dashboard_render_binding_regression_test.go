package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestDashboardRendersTodayEntryPeriodAndNotesFromStoredLog(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "dashboard-render-binding@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	today := dateAtLocation(time.Now().In(time.UTC), time.UTC)
	entry := models.DailyLog{
		UserID:   user.ID,
		Date:     today,
		IsPeriod: true,
		Flow:     models.FlowMedium,
		Notes:    "Stored dashboard note binding check",
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
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

	if !strings.Contains(rendered, "Stored dashboard note binding check") {
		t.Fatalf("expected dashboard textarea to include stored notes")
	}
	periodCheckedPattern := regexp.MustCompile(`(?s)name="is_period"[^>]*checked`)
	if !periodCheckedPattern.MatchString(rendered) {
		t.Fatalf("expected period toggle to include checked attribute for stored period day")
	}
	flowMediumPattern := regexp.MustCompile(`(?s)name="flow"[^>]*value="medium"[^>]*checked`)
	if !flowMediumPattern.MatchString(rendered) {
		t.Fatalf("expected medium flow option to remain checked from stored entry")
	}
}
