package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestCalendarDayPanelShowsDeleteForLegacyTimestampedLog(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-legacy@example.com", "StrongPass1", true)

	legacyTimestamp := "2026-01-15T15:30:00Z"
	now := time.Now().UTC()
	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		legacyTimestamp,
		true,
		models.FlowMedium,
		"[]",
		"legacy entry",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert legacy log: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	panelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-01-15", nil)
	panelRequest.Header.Set("Cookie", authCookie)
	panelResponse, err := app.Test(panelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer panelResponse.Body.Close()

	if panelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", panelResponse.StatusCode)
	}

	panelBody, err := io.ReadAll(panelResponse.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	if !strings.Contains(string(panelBody), "/api/log/delete?date=2026-01-15&source=calendar") {
		t.Fatalf("expected delete button for legacy timestamped log")
	}

	deleteRequest := httptest.NewRequest(http.MethodDelete, "/api/log/delete?date=2026-01-15&source=calendar", nil)
	deleteRequest.Header.Set("Cookie", authCookie)
	deleteResponse, err := app.Test(deleteRequest, -1)
	if err != nil {
		t.Fatalf("delete day request failed: %v", err)
	}
	defer deleteResponse.Body.Close()

	if deleteResponse.StatusCode != http.StatusNoContent {
		t.Fatalf("expected delete status 204, got %d", deleteResponse.StatusCode)
	}

	var remaining int64
	dayStart := time.Date(2026, time.January, 15, 0, 0, 0, 0, time.UTC)
	dayEnd := dayStart.AddDate(0, 0, 1)
	if err := database.Model(&models.DailyLog{}).
		Where("user_id = ? AND date >= ? AND date < ?", user.ID, dayStart, dayEnd).
		Count(&remaining).Error; err != nil {
		t.Fatalf("count remaining legacy logs: %v", err)
	}
	if remaining != 0 {
		t.Fatalf("expected legacy day logs to be deleted, got %d", remaining)
	}
}
