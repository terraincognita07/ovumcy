package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestCalendarDayExistsAndDeleteButtonUseAnyLegacyRowInDayRange(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-day-range@example.com", "StrongPass1", true)

	now := time.Now().UTC()
	dayWithData := "2026-02-17T08:30:00Z"
	dayWithoutData := "2026-02-17T20:15:00Z"

	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		dayWithData,
		true,
		models.FlowMedium,
		"[]",
		"has data",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert legacy data row: %v", err)
	}

	if err := database.Exec(
		`INSERT INTO daily_logs (user_id, date, is_period, flow, symptom_ids, notes, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		user.ID,
		dayWithoutData,
		false,
		models.FlowNone,
		"[]",
		"",
		now,
		now,
	).Error; err != nil {
		t.Fatalf("insert legacy empty row: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	existsRequest := httptest.NewRequest(http.MethodGet, "/api/days/2026-02-17/exists", nil)
	existsRequest.Header.Set("Accept", "application/json")
	existsRequest.Header.Set("Cookie", authCookie)
	existsResponse, err := app.Test(existsRequest, -1)
	if err != nil {
		t.Fatalf("exists request failed: %v", err)
	}
	defer existsResponse.Body.Close()

	if existsResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected exists status 200, got %d", existsResponse.StatusCode)
	}

	existsPayload := map[string]bool{}
	existsBody, err := io.ReadAll(existsResponse.Body)
	if err != nil {
		t.Fatalf("read exists body: %v", err)
	}
	if err := json.Unmarshal(existsBody, &existsPayload); err != nil {
		t.Fatalf("decode exists body: %v", err)
	}
	if !existsPayload["exists"] {
		t.Fatalf("expected exists=true when any row in day range has data")
	}

	panelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	panelRequest.Header.Set("Cookie", authCookie)
	panelResponse, err := app.Test(panelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer panelResponse.Body.Close()

	if panelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected panel status 200, got %d", panelResponse.StatusCode)
	}

	panelBody, err := io.ReadAll(panelResponse.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	if !strings.Contains(string(panelBody), "/api/log/delete?date=2026-02-17&source=calendar") {
		t.Fatalf("expected delete button when any legacy row in day range has data")
	}
}
