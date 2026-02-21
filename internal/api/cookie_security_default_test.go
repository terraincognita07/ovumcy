package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSecureCookiesDisabledByDefault(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "insecure-cookies@example.com", "StrongPass1", true)

	loginForm := url.Values{
		"email":    {user.Email},
		"password": {"StrongPass1"},
	}
	loginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(loginForm.Encode()))
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	loginResponse, err := app.Test(loginRequest, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer loginResponse.Body.Close()

	authCookie := responseCookie(loginResponse.Cookies(), authCookieName)
	if authCookie == nil {
		t.Fatal("expected auth cookie on valid login")
	}
	if authCookie.Secure {
		t.Fatal("expected auth cookie Secure=false when COOKIE_SECURE is disabled")
	}

	languageRequest := httptest.NewRequest(http.MethodGet, "/lang/en?next=/login", nil)
	languageResponse, err := app.Test(languageRequest, -1)
	if err != nil {
		t.Fatalf("language switch request failed: %v", err)
	}
	defer languageResponse.Body.Close()

	languageCookie := responseCookie(languageResponse.Cookies(), languageCookieName)
	if languageCookie == nil {
		t.Fatal("expected language cookie on language switch")
	}
	if languageCookie.Secure {
		t.Fatal("expected language cookie Secure=false when COOKIE_SECURE is disabled")
	}
}
