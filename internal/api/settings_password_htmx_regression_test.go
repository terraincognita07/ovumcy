package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSettingsChangePasswordFormUsesHTMXInlineFeedback(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-password-htmx@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/settings", nil)
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("settings request failed: %v", err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read settings body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `hx-post="/api/settings/change-password"`) {
		t.Fatalf("expected change password form to submit via htmx")
	}
	if !strings.Contains(rendered, `hx-target="#settings-change-password-status"`) {
		t.Fatalf("expected inline feedback target for change password form")
	}
	if !strings.Contains(rendered, `id="settings-change-password-status"`) {
		t.Fatalf("expected dedicated change password feedback container")
	}
}
