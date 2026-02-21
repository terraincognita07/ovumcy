package api

import (
	"net/url"
	"testing"
)

func TestSettingsPasswordChangeUsesFlashSuccessOnRedirect(t *testing.T) {
	form := url.Values{
		"current_password": {"StrongPass1"},
		"new_password":     {"EvenStronger2"},
		"confirm_password": {"EvenStronger2"},
	}
	assertSettingsFlashSuccessScenario(t, "/api/settings/change-password", form, "Password changed successfully.")
}
