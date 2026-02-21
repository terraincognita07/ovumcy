package api

import (
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestBuildPrivacyMetaDescriptionFallback(t *testing.T) {
	t.Parallel()

	got := buildPrivacyMetaDescription(map[string]string{})
	if got != "Lume Privacy Policy - Zero data collection, self-hosted period tracker." {
		t.Fatalf("unexpected fallback description: %q", got)
	}
}

func TestBuildPrivacyPageDataGuestUsesLoginBackFallback(t *testing.T) {
	t.Parallel()

	data := buildPrivacyPageData(map[string]string{}, "https://evil.example/path", nil)
	if backPath, ok := data["BackPath"].(string); !ok || backPath != "/login" {
		t.Fatalf("expected guest back path /login, got %#v", data["BackPath"])
	}
	if _, exists := data["CurrentUser"]; exists {
		t.Fatalf("did not expect CurrentUser for guest payload")
	}
}

func TestBuildPrivacyPageDataAuthenticatedUsesDashboardBackFallback(t *testing.T) {
	t.Parallel()

	user := &models.User{Email: "privacy@example.com"}
	data := buildPrivacyPageData(map[string]string{}, "https://evil.example/path", user)

	if backPath, ok := data["BackPath"].(string); !ok || backPath != "/dashboard" {
		t.Fatalf("expected auth back path /dashboard, got %#v", data["BackPath"])
	}
	currentUser, ok := data["CurrentUser"].(*models.User)
	if !ok || currentUser != user {
		t.Fatalf("expected CurrentUser pointer to be preserved, got %#v", data["CurrentUser"])
	}
}
