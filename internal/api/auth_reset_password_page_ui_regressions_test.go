package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

func TestResetPasswordPageShowsPasswordTogglesAndBackToLoginLink(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "reset-page-ui@example.com", "StrongPass1", true)

	recoveryCode, recoveryHash, err := generateRecoveryCodeHash()
	if err != nil {
		t.Fatalf("generate recovery code: %v", err)
	}
	if err := database.Model(&models.User{}).Where("id = ?", user.ID).Update("recovery_code_hash", recoveryHash).Error; err != nil {
		t.Fatalf("update recovery hash: %v", err)
	}

	forgotForm := url.Values{
		"recovery_code": {recoveryCode},
	}
	forgotRequest := httptest.NewRequest(http.MethodPost, "/api/auth/forgot-password", strings.NewReader(forgotForm.Encode()))
	forgotRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	forgotResponse, err := app.Test(forgotRequest, -1)
	if err != nil {
		t.Fatalf("forgot-password request failed: %v", err)
	}
	defer forgotResponse.Body.Close()

	if forgotResponse.StatusCode != http.StatusSeeOther {
		t.Fatalf("expected forgot-password status 303, got %d", forgotResponse.StatusCode)
	}
	location := forgotResponse.Header.Get("Location")
	if !strings.HasPrefix(location, "/reset-password?token=") {
		t.Fatalf("expected redirect to reset password page with token, got %q", location)
	}

	resetRequest := httptest.NewRequest(http.MethodGet, location, nil)
	resetResponse, err := app.Test(resetRequest, -1)
	if err != nil {
		t.Fatalf("reset-password page request failed: %v", err)
	}
	defer resetResponse.Body.Close()

	if resetResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected reset-password page status 200, got %d", resetResponse.StatusCode)
	}

	body, err := io.ReadAll(resetResponse.Body)
	if err != nil {
		t.Fatalf("read reset-password body: %v", err)
	}
	rendered := string(body)
	if strings.Count(rendered, `data-password-toggle`) < 2 {
		t.Fatalf("expected password toggle buttons on both reset password fields")
	}
	if !strings.Contains(rendered, `href="/login"`) {
		t.Fatalf("expected back-to-login link on reset password page")
	}
}
