package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRegisterPageKeepsEmailAfterPasswordValidationError(t *testing.T) {
	app, _ := newOnboardingTestApp(t)
	email := "persist-register@example.com"

	form := url.Values{
		"email":            {email},
		"password":         {"12345678"},
		"confirm_password": {"12345678"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("register request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	location := response.Header.Get("Location")
	if location == "" {
		t.Fatalf("expected redirect location after register validation error")
	}
	parsedLocation, err := url.Parse(location)
	if err != nil {
		t.Fatalf("parse redirect location: %v", err)
	}
	if parsedLocation.Path != "/register" {
		t.Fatalf("expected redirect to /register, got %q", parsedLocation.Path)
	}
	if parsedLocation.Query().Get("error") != "weak password" {
		t.Fatalf("expected weak password query fallback, got %q", parsedLocation.Query().Get("error"))
	}
	if parsedLocation.Query().Get("email") != email {
		t.Fatalf("expected email query fallback %q, got %q", email, parsedLocation.Query().Get("email"))
	}

	registerRequest := httptest.NewRequest(http.MethodGet, location, nil)
	registerResponse, err := app.Test(registerRequest, -1)
	if err != nil {
		t.Fatalf("register page request failed: %v", err)
	}
	defer registerResponse.Body.Close()

	if registerResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", registerResponse.StatusCode)
	}

	body, err := io.ReadAll(registerResponse.Body)
	if err != nil {
		t.Fatalf("read register body: %v", err)
	}
	if !strings.Contains(string(body), `id="register-email" type="email" name="email" value="`+email+`"`) {
		t.Fatalf("expected register page to keep submitted email after validation error")
	}
}
