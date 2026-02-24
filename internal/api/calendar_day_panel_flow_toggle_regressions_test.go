package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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
	if !strings.Contains(rendered, `data-day-editor-autosave="true"`) {
		t.Fatalf("expected calendar day editor form to enable autosave hooks")
	}
	if strings.Count(rendered, `:disabled="!isPeriod"`) != 1 {
		t.Fatalf("expected only flow controls to depend on period toggle")
	}
	if !strings.Contains(rendered, `name="symptom_ids"`) {
		t.Fatalf("expected symptoms controls to stay available regardless of period toggle")
	}
	if !strings.Contains(rendered, "All fields are auto-saved") {
		t.Fatalf("expected autosave hint text in day editor panel")
	}
}
