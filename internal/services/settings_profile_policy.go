package services

import (
	"errors"
	"strings"
	"unicode/utf8"
)

const maxSettingsDisplayNameLength = 64

var ErrSettingsDisplayNameTooLong = errors.New("settings display name too long")

func (service *SettingsService) NormalizeDisplayName(raw string) (string, error) {
	displayName := strings.TrimSpace(raw)
	if utf8.RuneCountInString(displayName) > maxSettingsDisplayNameLength {
		return "", ErrSettingsDisplayNameTooLong
	}
	return displayName, nil
}

func (service *SettingsService) ResolveProfileUpdateStatus(previousDisplayName string, updatedDisplayName string) string {
	status := "profile_updated"
	if strings.TrimSpace(previousDisplayName) != "" && updatedDisplayName == "" {
		status = "profile_name_cleared"
	}
	return status
}
