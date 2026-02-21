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

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	authCookie := responseCookieValue(response.Cookies(), authCookieName)
	if authCookie == "" {
		t.Fatalf("expected auth cookie in register response for auto-login")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, "Save your recovery code") {
		t.Fatalf("expected recovery code screen after register")
	}
	if !strings.Contains(rendered, "Continue to app") {
		t.Fatalf("expected continue button after register")
	}
}
