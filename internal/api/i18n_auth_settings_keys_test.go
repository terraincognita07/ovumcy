package api

import "testing"

func TestAuthErrorTranslationKey_NormalizesInput(t *testing.T) {
	got := authErrorTranslationKey("  TOO MANY LOGIN ATTEMPTS ")
	want := "auth.error.too_many_login_attempts"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}

	if got := authErrorTranslationKey(" PERIOD LENGTH MUST NOT EXCEED CYCLE LENGTH "); got != "onboarding.error.period_length_exceeds_cycle" {
		t.Fatalf("expected onboarding period-length/cycle-length key, got %q", got)
	}
	if got := authErrorTranslationKey(" PERIOD LENGTH IS INCOMPATIBLE WITH CYCLE LENGTH "); got != "settings.cycle.error_incompatible" {
		t.Fatalf("expected settings cycle compatibility key, got %q", got)
	}
}

func TestSettingsStatusTranslationKey(t *testing.T) {
	if got := settingsStatusTranslationKey("  CYCLE_UPDATED "); got != "settings.success.cycle_updated" {
		t.Fatalf("expected cycle_updated key, got %q", got)
	}
	if got := settingsStatusTranslationKey("unknown"); got != "" {
		t.Fatalf("expected empty key for unknown status, got %q", got)
	}
}
