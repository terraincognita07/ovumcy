package api

import "testing"

func TestPickFirstNonEmpty(t *testing.T) {
	got := pickFirstNonEmpty("   ", "", "  value ", "fallback")
	if got != "value" {
		t.Fatalf("expected first non-empty trimmed value, got %q", got)
	}
}

func TestClassifySettingsErrorSource_ChangePasswordError(t *testing.T) {
	errorKey, changePasswordErrorKey := classifySettingsErrorSource("password mismatch")
	if errorKey != "" {
		t.Fatalf("expected empty general error key, got %q", errorKey)
	}
	if changePasswordErrorKey != "auth.error.password_mismatch" {
		t.Fatalf("expected change-password error key, got %q", changePasswordErrorKey)
	}
}

func TestClassifySettingsErrorSource_GeneralError(t *testing.T) {
	errorKey, changePasswordErrorKey := classifySettingsErrorSource("invalid profile input")
	if errorKey != "settings.error.invalid_profile_input" {
		t.Fatalf("expected general error key, got %q", errorKey)
	}
	if changePasswordErrorKey != "" {
		t.Fatalf("expected empty change-password error key, got %q", changePasswordErrorKey)
	}
}

func TestClassifySettingsErrorSource_UnknownError(t *testing.T) {
	errorKey, changePasswordErrorKey := classifySettingsErrorSource("unknown error")
	if errorKey != "" || changePasswordErrorKey != "" {
		t.Fatalf("expected no translation keys for unknown error, got error=%q change_password=%q", errorKey, changePasswordErrorKey)
	}
}
