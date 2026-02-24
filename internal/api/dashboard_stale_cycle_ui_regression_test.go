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

func TestDashboardStaleCycleWarningIncludesSettingsCTAAndEstimatedPhase(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "dashboard-stale-ui@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	lastPeriodStart := dateAtLocation(time.Now().UTC(), time.UTC).AddDate(0, 0, -60)
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

	if !strings.Contains(rendered, "Cycle data may be outdated.") {
		t.Fatalf("expected stale cycle warning in dashboard")
	}
	if !strings.Contains(rendered, `href="/settings#settings-cycle"`) {
		t.Fatalf("expected stale cycle warning to include direct settings CTA")
	}
	if !strings.Contains(rendered, "Unknown") {
		t.Fatalf("expected phase to be rendered as unknown while cycle data is stale")
	}
}
