package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"
)

func TestSettingsChangePasswordInvalidCurrentPasswordShowsTopErrorBanner(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "settings-password-error@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"current_password": {"WrongPass1"},
		"new_password":     {"EvenStronger2"},
		"confirm_password": {"EvenStronger2"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/settings/change-password", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept-Language", "en")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("change-password request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/settings" {
		t.Fatalf("expected redirect to /settings, got %q", location)
	}

	flashValue := responseCookieValue(response.Cookies(), flashCookieName)
	if flashValue == "" {
		t.Fatalf("expected flash cookie for invalid current password")
	}

	followRequest := httptest.NewRequest(http.MethodGet, "/settings", nil)
	followRequest.Header.Set("Accept-Language", "en")
	followRequest.Header.Set("Cookie", authCookie+"; "+flashCookieName+"="+flashValue)
	followResponse, err := app.Test(followRequest, -1)
	if err != nil {
		t.Fatalf("follow-up settings request failed: %v", err)
	}
	defer followResponse.Body.Close()

	body, err := io.ReadAll(followResponse.Body)
	if err != nil {
		t.Fatalf("read settings body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, "Current password is incorrect.") {
		t.Fatalf("expected localized invalid-current-password banner")
	}

	topBannerPattern := regexp.MustCompile(`(?s)<section class="space-y-6">.*?Current password is incorrect\..*?<section class="journal-card p-5 sm:p-6" id="settings-change-password">`)
	if !topBannerPattern.MatchString(rendered) {
		t.Fatalf("expected invalid-current-password message in top settings banner area")
	}
}
