package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRegisterSuccessSetsAuthCookieAndShowsRecoveryStep(t *testing.T) {
	app, _ := newOnboardingTestApp(t)
	email := "autologin-register@example.com"

	form := url.Values{
		"email":            {email},
		"password":         {"StrongPass1"},
		"confirm_password": {"StrongPass1"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept-Language", "en")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register success request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/recovery-code" {
		t.Fatalf("expected redirect to /recovery-code, got %q", location)
	}

	authCookie := responseCookieValue(response.Cookies(), authCookieName)
	if authCookie == "" {
		t.Fatalf("expected auth cookie in register response for auto-login")
	}
	recoveryCookie := responseCookieValue(response.Cookies(), recoveryCodeCookieName)
	if recoveryCookie == "" {
		t.Fatalf("expected recovery page cookie in register response")
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
		t.Fatalf("expected recovery page status 200, got %d", recoveryResponse.StatusCode)
	}

	body, err := io.ReadAll(recoveryResponse.Body)
	if err != nil {
		t.Fatalf("read recovery response body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, "Save your recovery code") {
		t.Fatalf("expected recovery code screen after register")
	}
	if !strings.Contains(rendered, "Continue to app") {
		t.Fatalf("expected continue button after register")
	}
}
