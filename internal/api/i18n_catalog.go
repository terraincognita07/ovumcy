package api

import (
	"strings"
)

func settingsStatusTranslationKey(status string) string {
	switch strings.ToLower(strings.TrimSpace(status)) {
	case "password_changed":
		return "settings.success.password_changed"
	case "cycle_updated":
		return "settings.success.cycle_updated"
	case "profile_updated":
		return "settings.success.profile_updated"
	case "profile_name_cleared":
		return "settings.success.profile_name_cleared"
	case "data_cleared":
		return "settings.success.data_cleared"
	default:
		return ""
	}
}
