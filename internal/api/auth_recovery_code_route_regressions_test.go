package api

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRecoveryCodePageRedirectsToDashboardWhenCookieMissing(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "recovery-route-missing-cookie@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	request := httptest.NewRequest(http.MethodGet, "/recovery-code", nil)
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("recovery-code request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %q", location)
	}
}

func TestRecoveryCodePageRejectsCookieFromDifferentUser(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	userA := createOnboardingTestUser(t, database, "recovery-cookie-user-a@example.com", "StrongPass1", true)
	userB := createOnboardingTestUser(t, database, "recovery-cookie-user-b@example.com", "StrongPass1", true)
	authCookieUserB := loginAndExtractAuthCookie(t, app, userB.Email, "StrongPass1")

	payload := recoveryCodePagePayload{
		UserID:       userA.ID,
		RecoveryCode: "LUME-TEST-CODE-1234",
		ContinuePath: "/dashboard",
	}
	serialized, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal recovery payload: %v", err)
	}
	encoded := base64.RawURLEncoding.EncodeToString(serialized)

	request := httptest.NewRequest(http.MethodGet, "/recovery-code", nil)
	request.Header.Set("Cookie", authCookieUserB+"; "+recoveryCodeCookieName+"="+encoded)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("recovery-code request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected status 303, got %d", response.StatusCode)
	}
	if location := response.Header.Get("Location"); location != "/dashboard" {
		t.Fatalf("expected redirect to /dashboard, got %q", location)
	}

	cleared := responseCookie(response.Cookies(), recoveryCodeCookieName)
	if cleared == nil {
		t.Fatalf("expected invalid recovery cookie to be cleared")
	}
	if cleared.Value != "" {
		t.Fatalf("expected cleared recovery cookie value, got %q", cleared.Value)
	}
}
