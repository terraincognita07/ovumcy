package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestDashboardAndStatsUseSameStalePhasePresentation(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "dashboard-stats-stale-phase@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	lastPeriodStart := dateAtLocation(time.Now().UTC(), time.UTC).AddDate(0, 0, -60)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":      28,
		"period_length":     5,
		"last_period_start": lastPeriodStart,
	}).Error; err != nil {
		t.Fatalf("update stale baseline for user: %v", err)
	}

	dashboardRequest := httptest.NewRequest(http.MethodGet, "/dashboard", nil)
	dashboardRequest.Header.Set("Accept-Language", "en")
	dashboardRequest.Header.Set("Cookie", authCookie)
	dashboardResponse, err := app.Test(dashboardRequest, -1)
	if err != nil {
		t.Fatalf("dashboard request failed: %v", err)
	}
	defer dashboardResponse.Body.Close()

	dashboardBody, err := io.ReadAll(dashboardResponse.Body)
	if err != nil {
		t.Fatalf("read dashboard body: %v", err)
	}
	if !strings.Contains(string(dashboardBody), "Unknown") {
		t.Fatalf("expected dashboard stale phase to render as unknown")
	}

	statsRequest := httptest.NewRequest(http.MethodGet, "/stats", nil)
	statsRequest.Header.Set("Accept-Language", "en")
	statsRequest.Header.Set("Cookie", authCookie)
	statsResponse, err := app.Test(statsRequest, -1)
	if err != nil {
		t.Fatalf("stats request failed: %v", err)
	}
	defer statsResponse.Body.Close()

	statsBody, err := io.ReadAll(statsResponse.Body)
	if err != nil {
		t.Fatalf("read stats body: %v", err)
	}
	renderedStats := string(statsBody)
	if !strings.Contains(renderedStats, "Unknown") {
		t.Fatalf("expected stats stale phase to render as unknown")
	}
	if !strings.Contains(renderedStats, "Phase shown as estimate due stale baseline.") {
		t.Fatalf("expected stats stale phase hint to be rendered")
	}
}
