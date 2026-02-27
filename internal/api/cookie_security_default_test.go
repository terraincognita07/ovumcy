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
	if !authCookie.HttpOnly {
		t.Fatal("expected auth cookie HttpOnly=true")
	}
	if authCookie.Secure {
		t.Fatal("expected auth cookie Secure=false when COOKIE_SECURE is disabled")
	}
	if authCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected auth cookie SameSite=Lax, got %v", authCookie.SameSite)
	}

	registerForm := url.Values{
		"email":            {"recovery-cookie-default@example.com"},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass1"},
	}
	registerRequest := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(registerForm.Encode()))
	registerRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	registerResponse, err := app.Test(registerRequest, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer registerResponse.Body.Close()

	if registerResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected register status 303, got %d", registerResponse.StatusCode)
	}

	recoveryCookie := responseCookie(registerResponse.Cookies(), recoveryCodeCookieName)
	if recoveryCookie == nil {
		t.Fatal("expected recovery cookie after successful register")
	}
	if !recoveryCookie.HttpOnly {
		t.Fatal("expected recovery cookie HttpOnly=true")
	}
	if recoveryCookie.Secure {
		t.Fatal("expected recovery cookie Secure=false when COOKIE_SECURE is disabled")
	}
	if recoveryCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected recovery cookie SameSite=Lax, got %v", recoveryCookie.SameSite)
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
