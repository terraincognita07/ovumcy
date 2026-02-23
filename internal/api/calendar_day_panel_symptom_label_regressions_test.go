package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

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
