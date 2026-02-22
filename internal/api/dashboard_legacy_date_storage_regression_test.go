package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
)

func TestDashboardRendersTodayEntryFromLegacyDateOnlyStorage(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "dashboard-legacy-date@example.com", "StrongPass1", true)

	symptom := models.SymptomType{UserID: user.ID, Name: "Legacy symptom", Icon: "L", Color: "#000000"}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	today := dateAtLocation(time.Now().In(time.UTC), time.UTC)
	entry := models.DailyLog{
		UserID:     user.ID,
		Date:       today,
		IsPeriod:   true,
		Flow:       models.FlowMedium,
		SymptomIDs: []uint{symptom.ID},
		Notes:      "legacy date storage note",
	}
	if err := database.Create(&entry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	if err := database.Exec("UPDATE daily_logs SET date = ? WHERE id = ?", today.Format("2006-01-02"), entry.ID).Error; err != nil {
		t.Fatalf("simulate legacy date-only storage: %v", err)
	}

	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

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

	if !strings.Contains(rendered, "legacy date storage note") {
		t.Fatalf("expected dashboard notes field to include stored note from legacy date format")
	}

	periodCheckedPattern := regexp.MustCompile(`(?s)name="is_period"[^>]*checked`)
	if !periodCheckedPattern.MatchString(rendered) {
		t.Fatalf("expected dashboard period toggle to remain checked for legacy date storage")
	}

	symptomCheckedPattern := regexp.MustCompile(`(?s)name="symptom_ids"[^>]*value="` + strconv.FormatUint(uint64(symptom.ID), 10) + `"[^>]*checked`)
	if !symptomCheckedPattern.MatchString(rendered) {
		t.Fatalf("expected stored symptom to remain checked for legacy date storage")
	}
}
