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

func TestCalendarDayPanelDeleteEntryUsesConfirmForm(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-confirm@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	logEntry := models.DailyLog{
		UserID:   user.ID,
		Date:     time.Date(2026, time.February, 17, 0, 0, 0, 0, time.UTC),
		IsPeriod: true,
		Flow:     models.FlowMedium,
		Notes:    "entry",
	}
	if err := database.Create(&logEntry).Error; err != nil {
		t.Fatalf("create daily log: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `hx-delete="/api/log/delete?date=2026-02-17&source=calendar"`) {
		t.Fatalf("expected delete endpoint in day panel")
	}
	if !strings.Contains(rendered, `data-confirm="Are you sure you want to delete this entry?"`) {
		t.Fatalf("expected confirm prompt on calendar delete entry action")
	}
	if !strings.Contains(rendered, `data-confirm-accept="Yes, delete"`) {
		t.Fatalf("expected confirm accept label on calendar delete entry action")
	}
}

func TestCalendarDayPanelFlowControlsDependOnPeriodToggle(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-flow-toggle@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	panelRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	panelRequest.Header.Set("Accept-Language", "en")
	panelRequest.Header.Set("Cookie", authCookie)

	panelResponse, err := app.Test(panelRequest, -1)
	if err != nil {
		t.Fatalf("calendar day panel request failed: %v", err)
	}
	defer panelResponse.Body.Close()

	if panelResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", panelResponse.StatusCode)
	}

	body, err := io.ReadAll(panelResponse.Body)
	if err != nil {
		t.Fatalf("read panel body: %v", err)
	}
	rendered := string(body)

	if !strings.Contains(rendered, `x-data='dayEditorForm({ isPeriod: false })'`) {
		t.Fatalf("expected calendar panel form to initialize period state")
	}
	if !strings.Contains(rendered, `x-model="isPeriod"`) {
		t.Fatalf("expected period toggle to drive alpine state")
	}
	if !strings.Contains(rendered, `x-cloak x-show="isPeriod" :disabled="!isPeriod"`) {
		t.Fatalf("expected flow fieldset to be shown/enabled only when period is selected")
	}
	if strings.Count(rendered, `:disabled="!isPeriod"`) != 1 {
		t.Fatalf("expected only flow controls to depend on period toggle")
	}
	if !strings.Contains(rendered, `name="symptom_ids"`) {
		t.Fatalf("expected symptoms controls to stay available regardless of period toggle")
	}
}

func TestCalendarDayPanelUsesLanguageSpecificSymptomLabelClass(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "calendar-symptom-label-class@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	symptom := models.SymptomType{
		UserID: user.ID,
		Name:   "Breast tenderness",
		Icon:   "💗",
		Color:  "#D98395",
	}
	if err := database.Create(&symptom).Error; err != nil {
		t.Fatalf("create symptom: %v", err)
	}

	enRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	enRequest.Header.Set("Accept-Language", "en")
	enRequest.Header.Set("Cookie", authCookie)
	enResponse, err := app.Test(enRequest, -1)
	if err != nil {
		t.Fatalf("english panel request failed: %v", err)
	}
	defer enResponse.Body.Close()

	enBody, err := io.ReadAll(enResponse.Body)
	if err != nil {
		t.Fatalf("read english panel body: %v", err)
	}
	if !strings.Contains(string(enBody), `symptom-label symptom-label-nowrap`) {
		t.Fatalf("expected nowrap symptom class for english labels")
	}

	ruRequest := httptest.NewRequest(http.MethodGet, "/calendar/day/2026-02-17", nil)
	ruRequest.Header.Set("Accept-Language", "ru")
	ruRequest.Header.Set("Cookie", authCookie)
	ruResponse, err := app.Test(ruRequest, -1)
	if err != nil {
		t.Fatalf("russian panel request failed: %v", err)
	}
	defer ruResponse.Body.Close()

	ruBody, err := io.ReadAll(ruResponse.Body)
	if err != nil {
		t.Fatalf("read russian panel body: %v", err)
	}
	if strings.Contains(string(ruBody), `symptom-label symptom-label-nowrap`) {
		t.Fatalf("did not expect nowrap class for russian labels")
	}
	if !strings.Contains(string(ruBody), `class="symptom-label">`) {
		t.Fatalf("expected default symptom label class for russian locale")
	}
}
