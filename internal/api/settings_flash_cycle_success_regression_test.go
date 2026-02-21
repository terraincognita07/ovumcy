package api

import (
	"net/url"
	"testing"
)

func TestSettingsCycleUpdateUsesFlashSuccessOnRedirect(t *testing.T) {
	form := url.Values{
		"cycle_length":     {"29"},
		"period_length":    {"6"},
		"auto_period_fill": {"true"},
	}
	assertSettingsFlashSuccessScenario(t, "/settings/cycle", form, "Cycle settings updated successfully.")
}
