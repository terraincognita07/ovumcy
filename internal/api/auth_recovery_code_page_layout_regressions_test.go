package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestRecoveryCodePageUsesStandardCheckboxLayout(t *testing.T) {
	app, _ := newOnboardingTestApp(t)

	form := url.Values{
		"email":            {"recovery-checkbox@example.com"},
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
		t.Fatalf("read recovery code page body: %v", err)
	}
	rendered := string(body)
	if !strings.Contains(rendered, `class="remember-checkbox"`) {
		t.Fatalf("expected recovery page checkbox to use standard remember-checkbox style")
	}
	if strings.Contains(rendered, `class="choice-input"`) {
		t.Fatalf("did not expect recovery page checkbox to use chip toggle markup")
	}
	if !strings.Contains(rendered, `data-copy-success-message="Recovery code copied."`) {
		t.Fatalf("expected recovery page copy success feedback message attribute")
	}
	if !strings.Contains(rendered, `x-show="copyFailed"`) {
		t.Fatalf("expected recovery page to render explicit copy failure feedback state")
	}
	if !strings.Contains(rendered, `href="/lang/ru?next=%2Frecovery-code"`) {
		t.Fatalf("expected language switch link to keep recovery-code GET route")
	}
}
