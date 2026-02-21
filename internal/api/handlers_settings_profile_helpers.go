package api

import (
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
)

const maxDisplayNameLength = 64

func normalizeDisplayName(raw string) (string, error) {
	displayName := strings.TrimSpace(raw)
	if utf8.RuneCountInString(displayName) > maxDisplayNameLength {
		return "", fiber.NewError(fiber.StatusBadRequest, "display name too long")
	}
	return displayName, nil
}

func profileUpdateStatus(previousDisplayName string, updatedDisplayName string) string {
	status := "profile_updated"
	if strings.TrimSpace(previousDisplayName) != "" && updatedDisplayName == "" {
		status = "profile_name_cleared"
	}
	return status
}
