package services

import "testing"

func TestResolveSettingsStatusPrefersFlashValue(t *testing.T) {
	service := NewNotificationService()

	got := service.ResolveSettingsStatus(" cycle_updated ", "profile_updated", "password_changed")
	if got != "cycle_updated" {
		t.Fatalf("expected cycle_updated, got %q", got)
	}
}

func TestResolveSettingsStatusFallsBackToQuery(t *testing.T) {
	service := NewNotificationService()

	got := service.ResolveSettingsStatus("   ", "", " password_changed ")
	if got != "password_changed" {
		t.Fatalf("expected password_changed, got %q", got)
	}
}

func TestResolveSettingsErrorSourcePrefersFlashValue(t *testing.T) {
	service := NewNotificationService()

	got := service.ResolveSettingsErrorSource(" invalid current password ", "weak password")
	if got != "invalid current password" {
		t.Fatalf("expected invalid current password, got %q", got)
	}
}

func TestClassifySettingsErrorSource_ChangePassword(t *testing.T) {
	service := NewNotificationService()

	got := service.ClassifySettingsErrorSource(" password mismatch ")
	if got != SettingsErrorTargetChangePassword {
		t.Fatalf("expected change-password target, got %q", got)
	}
}

func TestClassifySettingsErrorSource_General(t *testing.T) {
	service := NewNotificationService()

	got := service.ClassifySettingsErrorSource("invalid profile input")
	if got != SettingsErrorTargetGeneral {
		t.Fatalf("expected general target, got %q", got)
	}
}
