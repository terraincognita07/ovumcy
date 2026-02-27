package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestSecureCookiesEnabledWhenConfigured(t *testing.T) {
	t.Parallel()

	app, database := newOnboardingTestAppWithCookieSecure(t, true)
	user := createOnboardingTestUser(t, database, "secure-cookies@example.com", "StrongPass1", true)

	invalidLoginForm := url.Values{
		"email":    {user.Email},
		"password": {"WrongPass1"},
	}
	invalidLoginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(invalidLoginForm.Encode()))
	invalidLoginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	invalidLoginResponse, err := app.Test(invalidLoginRequest, -1)
	if err != nil {
		t.Fatalf("invalid login request failed: %v", err)
	}
	defer invalidLoginResponse.Body.Close()

	if invalidLoginResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected invalid login status 303, got %d", invalidLoginResponse.StatusCode)
	}

	flashCookie := responseCookie(invalidLoginResponse.Cookies(), flashCookieName)
	if flashCookie == nil {
		t.Fatal("expected flash cookie on invalid login")
	}
	if !flashCookie.Secure {
		t.Fatal("expected flash cookie Secure=true when COOKIE_SECURE is enabled")
	}

	languageRequest := httptest.NewRequest(http.MethodGet, "/lang/en?next=/login", nil)
	languageResponse, err := app.Test(languageRequest, -1)
	if err != nil {
		t.Fatalf("language switch request failed: %v", err)
	}
	defer languageResponse.Body.Close()

	if languageResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected language switch status 303, got %d", languageResponse.StatusCode)
	}

	languageCookie := responseCookie(languageResponse.Cookies(), languageCookieName)
	if languageCookie == nil {
		t.Fatal("expected language cookie on language switch")
	}
	if !languageCookie.Secure {
		t.Fatal("expected language cookie Secure=true when COOKIE_SECURE is enabled")
	}

	validLoginForm := url.Values{
		"email":    {user.Email},
		"password": {"StrongPass1"},
	}
	validLoginRequest := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(validLoginForm.Encode()))
	validLoginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	validLoginResponse, err := app.Test(validLoginRequest, -1)
	if err != nil {
		t.Fatalf("valid login request failed: %v", err)
	}
	defer validLoginResponse.Body.Close()

	if validLoginResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected valid login status 303, got %d", validLoginResponse.StatusCode)
	}

	authCookie := responseCookie(validLoginResponse.Cookies(), authCookieName)
	if authCookie == nil {
		t.Fatal("expected auth cookie on valid login")
	}
	if !authCookie.HttpOnly {
		t.Fatal("expected auth cookie HttpOnly=true")
	}
	if !authCookie.Secure {
		t.Fatal("expected auth cookie Secure=true when COOKIE_SECURE is enabled")
	}
	if authCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected auth cookie SameSite=Lax, got %v", authCookie.SameSite)
	}

	registerForm := url.Values{
		"email":            {"recovery-cookie-secure@example.com"},
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
	if !recoveryCookie.Secure {
		t.Fatal("expected recovery cookie Secure=true when COOKIE_SECURE is enabled")
	}
	if recoveryCookie.SameSite != http.SameSiteLaxMode {
		t.Fatalf("expected recovery cookie SameSite=Lax, got %v", recoveryCookie.SameSite)
	}
}
