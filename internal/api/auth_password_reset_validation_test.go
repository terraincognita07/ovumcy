package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
)

func TestChangePasswordRejectsWeakNumericPassword(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "change-password@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"current_password": {"StrongPass1"},
		"new_password":     {"12345678"},
		"confirm_password": {"12345678"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/settings/change-password", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("change password request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "weak password" {
		t.Fatalf("expected weak password error, got %q", errorValue)
	}

	var updatedUser models.User
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("StrongPass1")) != nil {
		t.Fatalf("expected old password hash to stay unchanged")
	}
	if bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("12345678")) == nil {
		t.Fatalf("expected weak password not to be applied")
	}
}

func TestChangePasswordRejectsPasswordMismatch(t *testing.T) {
	app, database := newOnboardingTestApp(t)
	user := createOnboardingTestUser(t, database, "change-password-mismatch@example.com", "StrongPass1", true)
	authCookie := loginAndExtractAuthCookie(t, app, user.Email, "StrongPass1")

	form := url.Values{
		"current_password": {"StrongPass1"},
		"new_password":     {"EvenStronger2"},
		"confirm_password": {"DifferentPass3"},
	}
	request := httptest.NewRequest(http.MethodPost, "/api/settings/change-password", strings.NewReader(form.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	request.Header.Set("Accept", "application/json")
	request.Header.Set("Cookie", authCookie)

	response, err := app.Test(request, -1)
	if err != nil {
		t.Fatalf("change password mismatch request failed: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", response.StatusCode)
	}

	errorValue := readAPIError(t, response.Body)
	if errorValue != "password mismatch" {
		t.Fatalf("expected password mismatch error, got %q", errorValue)
	}

	var updatedUser models.User
	if err := database.First(&updatedUser, user.ID).Error; err != nil {
		t.Fatalf("load updated user: %v", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(updatedUser.PasswordHash), []byte("StrongPass1")) != nil {
		t.Fatalf("expected old password hash to stay unchanged")
	}
}

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
