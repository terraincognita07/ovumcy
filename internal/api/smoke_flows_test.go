package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func TestOwnerCriticalFlowSmoke(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "smoke-owner@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	smokeGET(t, app, authCookie, "/dashboard", http.StatusOK)

	settingsBody := smokeGET(t, app, authCookie, "/settings", http.StatusOK)
	if !strings.Contains(settingsBody, `data-export-section`) {
		t.Fatalf("expected settings page to include export section")
	}

	calendarBody := smokeGET(t, app, authCookie, "/calendar?month=2026-02", http.StatusOK)
	if !strings.Contains(calendarBody, `id="day-editor"`) {
		t.Fatalf("expected calendar page to render day editor target")
	}

	payload := map[string]any{
		"is_period":   true,
		"flow":        models.FlowMedium,
		"symptom_ids": []uint{},
		"notes":       "smoke flow note",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal day payload: %v", err)
	}

	upsertRequest := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-21", bytes.NewReader(body))
	upsertRequest.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	upsertRequest.Header.Set("Cookie", authCookie)

	upsertResponse, err := app.Test(upsertRequest, -1)
	if err != nil {
		t.Fatalf("upsert day request failed: %v", err)
	}
	defer upsertResponse.Body.Close()

	if upsertResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected upsert status 200, got %d", upsertResponse.StatusCode)
	}

	dayPanelBody := smokeGET(t, app, authCookie, "/calendar/day/2026-02-21", http.StatusOK)
	if !strings.Contains(dayPanelBody, "smoke flow note") {
		t.Fatalf("expected day panel to include persisted day notes")
	}

	exportRequest := httptest.NewRequest(http.MethodGet, "/api/export/csv?from=2026-02-01&to=2026-02-28", nil)
	exportRequest.Header.Set("Cookie", authCookie)

	exportResponse, err := app.Test(exportRequest, -1)
	if err != nil {
		t.Fatalf("export request failed: %v", err)
	}
	defer exportResponse.Body.Close()

	if exportResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected export status 200, got %d", exportResponse.StatusCode)
	}
	if got := exportResponse.Header.Get("Content-Type"); !strings.Contains(got, "text/csv") {
		t.Fatalf("expected export content type text/csv, got %q", got)
	}

	exportBody, err := io.ReadAll(exportResponse.Body)
	if err != nil {
		t.Fatalf("read export body: %v", err)
	}
	renderedExport := string(exportBody)
	if !strings.Contains(renderedExport, "2026-02-21") {
		t.Fatalf("expected exported csv to include saved day date")
	}
	if !strings.Contains(renderedExport, "smoke flow note") {
		t.Fatalf("expected exported csv to include saved day notes")
	}
}

func TestPartnerReadOnlyFlowSmoke(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "smoke-partner@example.com", "StrongPass1", true)
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Update("role", models.RolePartner).Error; err != nil {
		t.Fatalf("set partner role: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.February, 21, 0, 0, 0, 0, time.UTC),
		IsPeriod:   true,
		Flow:       models.FlowMedium,
		SymptomIDs: []uint{1, 2},
		Notes:      "partner-private-note",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create partner log: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	dashboardBody := smokeGET(t, app, authCookie, "/dashboard", http.StatusOK)
	if strings.Contains(dashboardBody, `hx-post="/api/days/`) {
		t.Fatalf("expected partner dashboard to be read-only")
	}

	settingsBody := smokeGET(t, app, authCookie, "/settings", http.StatusOK)
	if strings.Contains(settingsBody, `data-export-section`) {
		t.Fatalf("expected owner-only export section hidden for partner")
	}
	if strings.Contains(settingsBody, `action="/settings/cycle"`) {
		t.Fatalf("expected owner-only cycle settings hidden for partner")
	}

	dayPanelBody := smokeGET(t, app, authCookie, "/calendar/day/2026-02-21", http.StatusOK)
	if strings.Contains(dayPanelBody, `hx-post="/api/days/2026-02-21"`) {
		t.Fatalf("expected partner day panel to be read-only")
	}
	if strings.Contains(dayPanelBody, `name="notes"`) {
		t.Fatalf("expected partner day notes field to be hidden")
	}
	if strings.Contains(dayPanelBody, `name="symptom_ids"`) {
		t.Fatalf("expected partner day symptoms controls to be hidden")
	}

	payload := map[string]any{
		"is_period":   true,
		"flow":        models.FlowMedium,
		"symptom_ids": []uint{},
		"notes":       "forbidden write",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal partner payload: %v", err)
	}

	upsertRequest := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-21", bytes.NewReader(body))
	upsertRequest.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	upsertRequest.Header.Set("Cookie", authCookie)

	upsertResponse, err := app.Test(upsertRequest, -1)
	if err != nil {
		t.Fatalf("partner upsert request failed: %v", err)
	}
	defer upsertResponse.Body.Close()

	if upsertResponse.StatusCode != http.StatusForbidden {
		t.Fatalf("expected partner upsert status 403, got %d", upsertResponse.StatusCode)
	}

	exportRequest := httptest.NewRequest(http.MethodGet, "/api/export/csv?from=2026-02-01&to=2026-02-28", nil)
	exportRequest.Header.Set("Cookie", authCookie)

	exportResponse, err := app.Test(exportRequest, -1)
	if err != nil {
		t.Fatalf("partner export request failed: %v", err)
	}
	defer exportResponse.Body.Close()

	if exportResponse.StatusCode != http.StatusForbidden {
		t.Fatalf("expected partner export status 403, got %d", exportResponse.StatusCode)
	}
}

func smokeGET(t *testing.T, app *fiber.App, authCookie string, path string, expectedStatus int) string {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("GET %s failed: %v", path, err)
	}
	defer response.Body.Close()

	if response.StatusCode != expectedStatus {
		t.Fatalf("GET %s expected status %d, got %d", path, expectedStatus, response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("GET %s read body failed: %v", path, err)
	}
	return string(body)
}
