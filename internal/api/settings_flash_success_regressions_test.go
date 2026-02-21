package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSettingsPageUsesFlashSuccessOnRedirect(t *testing.T) {
	testCases := []struct {
		name         string
		path         string
		form         url.Values
		successLabel string
	}{
		{
			name: "cycle settings success",
			path: "/settings/cycle",
			form: url.Values{
				"cycle_length":     {"29"},
				"period_length":    {"6"},
				"auto_period_fill": {"true"},
			},
			successLabel: "Cycle settings updated successfully.",
		},
		{
			name: "password status success",
			path: "/api/settings/change-password",
			form: url.Values{
				"current_password": {"StrongPass1"},
				"new_password":     {"EvenStronger2"},
				"confirm_password": {"EvenStronger2"},
			},
			successLabel: "Password changed successfully.",
		},
		{
			name:         "clear data success",
			path:         "/api/settings/clear-data",
			form:         url.Values{},
			successLabel: "All tracking data cleared successfully.",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			app, database := newOnboardingTestApp(t)
			user := createOnboardingTestUser(t, database, "settings-user@example.com", "StrongPass1", true)
			authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

			request := httptest.NewRequest(http.MethodPost, testCase.path, strings.NewReader(testCase.form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			request.Header.Set("Cookie", authCookie)

			response, err := app.Test(request, -1)
			if err != nil {
				t.Fatalf("settings action request failed: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusSeeOther {
				t.Fatalf("expected status 303, got %d", response.StatusCode)
			}
			if location := response.Header.Get("Location"); location != "/settings" {
				t.Fatalf("expected redirect %q, got %q", "/settings", location)
			}

			flashValue := responseCookieValue(response.Cookies(), flashCookieName)
			if flashValue == "" {
				t.Fatalf("expected flash cookie for settings success message")
			}

			followRequest := httptest.NewRequest(http.MethodGet, "/settings", nil)
			followRequest.Header.Set("Accept-Language", "en")
			followRequest.Header.Set("Cookie", authCookie+"; "+flashCookieName+"="+flashValue)

			followResponse, err := app.Test(followRequest, -1)
			if err != nil {
				t.Fatalf("follow-up settings request failed: %v", err)
			}
			defer followResponse.Body.Close()

			if followResponse.StatusCode != http.StatusOK {
				t.Fatalf("expected follow-up status 200, got %d", followResponse.StatusCode)
			}

			body, err := io.ReadAll(followResponse.Body)
			if err != nil {
				t.Fatalf("read follow-up body: %v", err)
			}
			rendered := string(body)
			if !strings.Contains(rendered, testCase.successLabel) {
				t.Fatalf("expected success label %q in settings page", testCase.successLabel)
			}
			if strings.Contains(rendered, weakPasswordErrorText) {
				t.Fatalf("did not expect weak password error on success page")
			}

			afterFlashRequest := httptest.NewRequest(http.MethodGet, "/settings", nil)
			afterFlashRequest.Header.Set("Accept-Language", "en")
			afterFlashRequest.Header.Set("Cookie", authCookie)

			afterFlashResponse, err := app.Test(afterFlashRequest, -1)
			if err != nil {
				t.Fatalf("settings request after flash consumption failed: %v", err)
			}
			defer afterFlashResponse.Body.Close()

			afterFlashBody, err := io.ReadAll(afterFlashResponse.Body)
			if err != nil {
				t.Fatalf("read settings body after flash consumption: %v", err)
			}
			if strings.Contains(string(afterFlashBody), testCase.successLabel) {
				t.Fatalf("did not expect success label after flash is consumed")
			}
		})
	}
}
