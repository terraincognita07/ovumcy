package api

import (
	"encoding/csv"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestExportCSVRespectsRequestedDateRange(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "export-range@example.com", "StrongPass1", true)

	entries := []models.DailyLog{
		{
			UserID:   user.ID,
			Date:     time.Date(2026, time.February, 2, 0, 0, 0, 0, time.UTC),
			IsPeriod: true,
			Flow:     models.FlowLight,
			Notes:    "before-range",
		},
		{
			UserID:   user.ID,
			Date:     time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC),
			IsPeriod: false,
			Flow:     models.FlowNone,
			Notes:    "in-range",
		},
		{
			UserID:   user.ID,
			Date:     time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC),
			IsPeriod: true,
			Flow:     models.FlowHeavy,
			Notes:    "after-range",
		},
	}
	if err := database.Create(&entries).Error; err != nil {
		t.Fatalf("create export logs: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	request := httptest.NewRequest(http.MethodGet, "/api/export/csv?from=2026-02-05&to=2026-02-12", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("export csv request with range failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	records, err := csv.NewReader(strings.NewReader(string(body))).ReadAll()
	if err != nil {
		t.Fatalf("parse csv: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected header + 1 row in selected range, got %d rows", len(records))
	}

	header := records[0]
	row := records[1]
	indexByName := make(map[string]int, len(header))
	for index, name := range header {
		indexByName[name] = index
	}

	if got := row[indexByName["Date"]]; got != "2026-02-10" {
		t.Fatalf("expected in-range date 2026-02-10, got %q", got)
	}
	if got := row[indexByName["Notes"]]; got != "in-range" {
		t.Fatalf("expected in-range notes, got %q", got)
	}
}

func TestExportJSONRejectsInvalidDateRange(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "export-invalid-range@example.com", "StrongPass1", true)

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	request := httptest.NewRequest(http.MethodGet, "/api/export/json?from=2026-02-20&to=2026-02-10", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("export json request with invalid range failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	payload := struct {
		Error string `json:"error"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload.Error != "invalid range" {
		t.Fatalf("expected invalid range error, got %q", payload.Error)
	}
}

func TestExportSummaryRespectsRequestedDateRange(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "export-summary@example.com", "StrongPass1", true)

	entries := []models.DailyLog{
		{
			UserID: user.ID,
			Date:   time.Date(2026, time.February, 7, 0, 0, 0, 0, time.UTC),
			Flow:   models.FlowNone,
		},
		{
			UserID: user.ID,
			Date:   time.Date(2026, time.February, 12, 0, 0, 0, 0, time.UTC),
			Flow:   models.FlowLight,
		},
		{
			UserID: user.ID,
			Date:   time.Date(2026, time.February, 20, 0, 0, 0, 0, time.UTC),
			Flow:   models.FlowHeavy,
		},
	}
	if err := database.Create(&entries).Error; err != nil {
		t.Fatalf("create export logs: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	request := httptest.NewRequest(http.MethodGet, "/api/export/summary?from=2026-02-10&to=2026-02-19", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("export summary request with range failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	payload := struct {
		TotalEntries int    `json:"total_entries"`
		HasData      bool   `json:"has_data"`
		DateFrom     string `json:"date_from"`
		DateTo       string `json:"date_to"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode summary payload: %v", err)
	}

	if payload.TotalEntries != 1 {
		t.Fatalf("expected total_entries 1, got %d", payload.TotalEntries)
	}
	if !payload.HasData {
		t.Fatal("expected has_data=true")
	}
	if payload.DateFrom != "2026-02-12" {
		t.Fatalf("expected date_from 2026-02-12, got %q", payload.DateFrom)
	}
	if payload.DateTo != "2026-02-12" {
		t.Fatalf("expected date_to 2026-02-12, got %q", payload.DateTo)
	}
}

func TestExportSummaryRejectsInvalidDateRange(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "export-summary-invalid-range@example.com", "StrongPass1", true)

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")
	request := httptest.NewRequest(http.MethodGet, "/api/export/summary?from=2026-02-20&to=2026-02-10", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("export summary request with invalid range failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}

	payload := struct {
		Error string `json:"error"`
	}{}
	if err := json.Unmarshal(body, &payload); err != nil {
		t.Fatalf("decode error response: %v", err)
	}
	if payload.Error != "invalid range" {
		t.Fatalf("expected invalid range error, got %q", payload.Error)
	}
}
