package api

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func TestUpsertDayNormalizesFlowWhenNotPeriod(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "upsert-day-normalize@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	payload := map[string]any{
		"is_period":   false,
		"flow":        models.FlowHeavy,
		"symptom_ids": []uint{},
		"notes":       "note",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-19", bytes.NewReader(body))
	request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("upsert request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	day, err := parseDayParam("2026-02-19", time.UTC)
	if err != nil {
		t.Fatalf("parse day for assertion: %v", err)
	}
	entry, err := (&Handler{db: database, location: time.UTC}).fetchLogByDate(user.ID, day)
	if err != nil {
		t.Fatalf("load stored log: %v", err)
	}
	if entry.Flow != models.FlowNone {
		t.Fatalf("expected non-period flow normalized to %q, got %q", models.FlowNone, entry.Flow)
	}
}

func TestUpsertDayRejectsPeriodWithoutFlow(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "upsert-day-invalid-flow@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	payload := map[string]any{
		"is_period":   true,
		"flow":        models.FlowNone,
		"symptom_ids": []uint{},
		"notes":       "note",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-19", bytes.NewReader(body))
	request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	request.Header.Set("Accept", fiber.MIMEApplicationJSON)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("upsert request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "period flow is required" {
		t.Fatalf("expected period flow validation error, got %q", errorValue)
	}
}

func TestUpsertDayAutoFillsFollowingPeriodDays(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "upsert-day-autofill@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"period_length":    4,
		"auto_period_fill": true,
	}).Error; err != nil {
		t.Fatalf("update user cycle settings: %v", err)
	}

	payload := map[string]any{
		"is_period":   true,
		"flow":        models.FlowMedium,
		"symptom_ids": []uint{},
		"notes":       "",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-10", bytes.NewReader(body))
	request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("upsert request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	handler := &Handler{db: database, location: time.UTC}
	autoFilledDays := []string{"2026-02-10", "2026-02-11", "2026-02-12", "2026-02-13"}
	for _, dateRaw := range autoFilledDays {
		day, err := parseDayParam(dateRaw, time.UTC)
		if err != nil {
			t.Fatalf("parse day %s: %v", dateRaw, err)
		}
		entry, err := handler.fetchLogByDate(user.ID, day)
		if err != nil {
			t.Fatalf("fetch log for %s: %v", dateRaw, err)
		}
		if !entry.IsPeriod {
			t.Fatalf("expected %s to be auto-marked as period day", dateRaw)
		}
	}

	dayAfterAutoFill, err := parseDayParam("2026-02-14", time.UTC)
	if err != nil {
		t.Fatalf("parse day after auto-fill: %v", err)
	}
	dayAfterEntry, err := handler.fetchLogByDate(user.ID, dayAfterAutoFill)
	if err != nil {
		t.Fatalf("fetch log for day after auto-fill: %v", err)
	}
	if dayAfterEntry.IsPeriod {
		t.Fatalf("expected day after auto-fill window to remain non-period")
	}
}

func TestUpsertDayAutoFillCanBeDisabled(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "upsert-day-autofill-disabled@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"period_length":    4,
		"auto_period_fill": false,
	}).Error; err != nil {
		t.Fatalf("update user cycle settings: %v", err)
	}

	payload := map[string]any{
		"is_period":   true,
		"flow":        models.FlowMedium,
		"symptom_ids": []uint{},
		"notes":       "",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-10", bytes.NewReader(body))
	request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("upsert request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	handler := &Handler{db: database, location: time.UTC}
	firstDay, err := parseDayParam("2026-02-10", time.UTC)
	if err != nil {
		t.Fatalf("parse first day: %v", err)
	}
	firstEntry, err := handler.fetchLogByDate(user.ID, firstDay)
	if err != nil {
		t.Fatalf("fetch first day log: %v", err)
	}
	if !firstEntry.IsPeriod {
		t.Fatalf("expected first selected day to be period")
	}

	nextDay, err := parseDayParam("2026-02-11", time.UTC)
	if err != nil {
		t.Fatalf("parse next day: %v", err)
	}
	nextEntry, err := handler.fetchLogByDate(user.ID, nextDay)
	if err != nil {
		t.Fatalf("fetch next day log: %v", err)
	}
	if nextEntry.IsPeriod {
		t.Fatalf("expected next day to stay manual when auto-fill is disabled")
	}
}

func TestUpsertDayAutoFillSkipsWhenRecentPeriodDayExists(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "upsert-day-autofill-recent-period@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"period_length":    4,
		"auto_period_fill": true,
	}).Error; err != nil {
		t.Fatalf("update user cycle settings: %v", err)
	}

	existingPeriodDay := time.Date(2026, time.February, 8, 0, 0, 0, 0, time.UTC)
	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       existingPeriodDay,
		IsPeriod:   true,
		Flow:       models.FlowMedium,
		SymptomIDs: []uint{},
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create existing period day: %v", err)
	}

	payload := map[string]any{
		"is_period":   true,
		"flow":        models.FlowMedium,
		"symptom_ids": []uint{},
		"notes":       "",
	}
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/api/days/2026-02-10", bytes.NewReader(body))
	request.Header.Set("Content-Type", fiber.MIMEApplicationJSON)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("upsert request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	handler := &Handler{db: database, location: time.UTC}
	nextDay, err := parseDayParam("2026-02-11", time.UTC)
	if err != nil {
		t.Fatalf("parse next day: %v", err)
	}
	nextEntry, err := handler.fetchLogByDate(user.ID, nextDay)
	if err != nil {
		t.Fatalf("fetch next day log: %v", err)
	}
	if nextEntry.IsPeriod {
		t.Fatalf("expected recent-period guard to prevent a new auto-fill sequence")
	}
}

func TestDeleteSymptomRemovesIDFromLogs(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "delete-symptom-cleanup@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Custom",
		Icon:   "A",
		Color:  "#111111",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	logEntry := models.DailyLog{
		UserID:     user.ID,
		Date:       time.Date(2026, time.February, 18, 0, 0, 0, 0, time.UTC),
		IsPeriod:   false,
		Flow:       models.FlowNone,
		SymptomIDs: []uint{symptom.ID, symptom.ID},
		Notes:      "",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	request := httptest.NewRequest(http.MethodDelete, "/api/symptoms/"+strconv.FormatUint(uint64(symptom.ID), 10), nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("delete symptom request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(response.Body)
		t.Fatalf("expected status 200, got %d: %s", response.StatusCode, string(body))
	}

	var deletedCount int64
	if err := database.Model(&models.SymptomType{}).Where("id = ?", symptom.ID).Count(&deletedCount).Error; err != nil {
		t.Fatalf("count deleted symptom: %v", err)
	}
	if deletedCount != 0 {
		t.Fatalf("expected symptom to be deleted, count=%d", deletedCount)
	}

	updated := models.DailyLog{}
	if err := database.First(&updated, logEntry.ID).Error; err != nil {
		t.Fatalf("load updated log: %v", err)
	}
	if len(updated.SymptomIDs) != 0 {
		t.Fatalf("expected symptom IDs to be removed from log, got %#v", updated.SymptomIDs)
	}
}
