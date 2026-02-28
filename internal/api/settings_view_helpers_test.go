package api

import (
	"testing"

	"github.com/terraincognita07/ovumcy/internal/services"
)

func TestResolveSettingsErrorKeys_ChangePasswordError(t *testing.T) {
	notificationService := services.NewNotificationService()
	errorKey, changePasswordErrorKey := resolveSettingsErrorKeys(notificationService, "password mismatch")
	if errorKey != "" {
		t.Fatalf("expected empty general error key, got %q", errorKey)
	}
	if changePasswordErrorKey != "auth.error.password_mismatch" {
		t.Fatalf("expected change-password error key, got %q", changePasswordErrorKey)
	}
}

func TestResolveSettingsErrorKeys_GeneralError(t *testing.T) {
	notificationService := services.NewNotificationService()
	errorKey, changePasswordErrorKey := resolveSettingsErrorKeys(notificationService, "invalid profile input")
	if errorKey != "settings.error.invalid_profile_input" {
		t.Fatalf("expected general error key, got %q", errorKey)
	}
	if changePasswordErrorKey != "" {
		t.Fatalf("expected empty change-password error key, got %q", changePasswordErrorKey)
	}
}

func TestResolveSettingsErrorKeys_UnknownError(t *testing.T) {
	notificationService := services.NewNotificationService()
	errorKey, changePasswordErrorKey := resolveSettingsErrorKeys(notificationService, "unknown error")
	if errorKey != "" || changePasswordErrorKey != "" {
		t.Fatalf("expected no translation keys for unknown error, got error=%q change_password=%q", errorKey, changePasswordErrorKey)
	}
}
