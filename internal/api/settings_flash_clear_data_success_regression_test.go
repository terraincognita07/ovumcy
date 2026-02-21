package api

import (
	"net/url"
	"testing"
)

func TestSettingsClearDataUsesFlashSuccessOnRedirect(t *testing.T) {
	assertSettingsFlashSuccessScenario(t, "/api/settings/clear-data", url.Values{}, "All tracking data cleared successfully.")
}
