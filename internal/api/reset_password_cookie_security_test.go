package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestResetPasswordCookieFlagsFollowCookieSecureConfig(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name             string
		cookieSecure     bool
		expectedSecure   bool
		expectedSameSite http.SameSite
	}{
		{
			name:             "cookie_secure_disabled",
			cookieSecure:     false,
			expectedSecure:   false,
			expectedSameSite: http.SameSiteLaxMode,
		},
		{
			name:             "cookie_secure_enabled",
			cookieSecure:     true,
			expectedSecure:   true,
			expectedSameSite: http.SameSiteLaxMode,
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			app, database := newOnboardingTestAppWithCookieSecure(t, tc.cookieSecure)
			user := createOnboardingTestUser(t, database, "reset-cookie-flags-"+tc.name+"@example.com", "StrongPass1", true)
			recoveryCode := mustSetRecoveryCodeForUser(t, database, user.ID)

			form := url.Values{"recovery_code": {recoveryCode}}
			request := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", strings.NewReader(form.Encode()))
			request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			response, err := app.Test(request, -1)
			if err != nil {
				t.Fatalf("forgot-password request failed: %v", err)
			}
			defer response.Body.Close()

			if response.StatusCode != http.StatusSeeOther {
				t.Fatalf("expected status 303, got %d", response.StatusCode)
			}

			resetCookie := responseCookie(response.Cookies(), resetPasswordCookieName)
			if resetCookie == nil {
				t.Fatalf("expected reset-password cookie in response")
			}
			if !resetCookie.HttpOnly {
				t.Fatalf("expected reset-password cookie HttpOnly=true")
			}
			if resetCookie.Secure != tc.expectedSecure {
				t.Fatalf("expected reset-password cookie Secure=%t, got %t", tc.expectedSecure, resetCookie.Secure)
			}
			if resetCookie.SameSite != tc.expectedSameSite {
				t.Fatalf("expected reset-password cookie SameSite=%v, got %v", tc.expectedSameSite, resetCookie.SameSite)
			}
		})
	}
}
