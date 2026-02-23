package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestAuthPagesIncludeSwitchTransitionHooks(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	loginRequest := httptest.NewRequest(http.MethodGet, "/login", nil)
	loginRequest.Header.Set("Accept-Language", "en")
	loginResponse, err := app.Test(loginRequest, -1)
	if err != nil {
		t.Fatalf("login request failed: %v", err)
	}
	defer loginResponse.Body.Close()

	loginBody, err := io.ReadAll(loginResponse.Body)
	if err != nil {
		t.Fatalf("read login body: %v", err)
	}
	loginRendered := string(loginBody)
	if !strings.Contains(loginRendered, `data-auth-panel`) {
		t.Fatalf("expected auth panel transition hook on login page")
	}
	if !strings.Contains(loginRendered, `data-auth-switch`) {
		t.Fatalf("expected auth switch transition hook on login page")
	}
	if !strings.Contains(loginRendered, `<script defer src="/static/js/app.js?v=`) {
		t.Fatalf("expected shared app script for auth panel transitions")
	}

	registerRequest := httptest.NewRequest(http.MethodGet, "/register", nil)
	registerRequest.Header.Set("Accept-Language", "en")
	registerResponse, err := app.Test(registerRequest, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer registerResponse.Body.Close()

	registerBody, err := io.ReadAll(registerResponse.Body)
	if err != nil {
		t.Fatalf("read register body: %v", err)
	}
	registerRendered := string(registerBody)
	if !strings.Contains(registerRendered, `data-auth-panel`) {
		t.Fatalf("expected auth panel transition hook on register page")
	}
	if !strings.Contains(registerRendered, `data-auth-switch`) {
		t.Fatalf("expected auth switch transition hook on register page")
	}
}

func TestRecoveryCodePageIncludesDownloadFeedbackMessage(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	form := url.Values{
		"email":            {"recovery-feedback@example.com"},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept-Language", "en")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/recovery-code" {
		t.Fatalf("expected redirect to /recovery-code, got %q", location)
	}

	authCookie := responseCookieValue(response.Cookies(), authCookieName)
	recoveryCookie := responseCookieValue(response.Cookies(), recoveryCodeCookieName)
	if authCookie == "" || recoveryCookie == "" {
		t.Fatalf("expected auth and recovery cookies in register response")
	}

	recoveryRequest := httptest.NewRequest(http.MethodGet, "/recovery-code", nil)
	recoveryRequest.Header.Set("Accept-Language", "en")
	recoveryRequest.Header.Set("Cookie", authCookieName+"="+authCookie+"; "+recoveryCodeCookieName+"="+recoveryCookie)

	recoveryResponse, err := app.Test(recoveryRequest, -1)
	if err != nil {
		t.Fatalf("recovery page request failed: %v", err)
	}
	defer recoveryResponse.Body.Close()

	if recoveryResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected recovery status 200, got %d", recoveryResponse.StatusCode)
	}

	body, err := io.ReadAll(recoveryResponse.Body)
	if err != nil {
		t.Fatalf("read recovery page body: %v", err)
	}
	if !strings.Contains(string(body), "Recovery code downloaded.") {
		t.Fatalf("expected recovery code download feedback message")
	}
}
