package api

import (
	"strings"
	"unicode/utf8"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"golang.org/x/crypto/bcrypt"
)

const maxDisplayNameLength = 64

func normalizeDisplayName(raw string) (string, error) {
	displayName := strings.TrimSpace(raw)
	if utf8.RuneCountInString(displayName) > maxDisplayNameLength {
		return "", fiber.NewError(fiber.StatusBadRequest, "display name too long")
	}
	return displayName, nil
}

func parseChangePasswordInput(c *fiber.Ctx) (changePasswordInput, string) {
	input := changePasswordInput{}
	if err := c.BodyParser(&input); err != nil {
		return changePasswordInput{}, "invalid settings input"
	}

	input.CurrentPassword = strings.TrimSpace(input.CurrentPassword)
	input.NewPassword = strings.TrimSpace(input.NewPassword)
	input.ConfirmPassword = strings.TrimSpace(input.ConfirmPassword)
	if input.CurrentPassword == "" || input.NewPassword == "" || input.ConfirmPassword == "" {
		return changePasswordInput{}, "invalid settings input"
	}
	if input.NewPassword != input.ConfirmPassword {
		return changePasswordInput{}, "password mismatch"
	}
	return input, ""
}

func validateChangePasswordInput(input changePasswordInput, user *models.User) (int, string) {
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.CurrentPassword)) != nil {
		return fiber.StatusUnauthorized, "invalid current password"
	}
	if input.CurrentPassword == input.NewPassword {
		return fiber.StatusBadRequest, "new password must differ"
	}
	if err := validatePasswordStrength(input.NewPassword); err != nil {
		return fiber.StatusBadRequest, "weak password"
	}
	return 0, ""
}

func parseCycleSettingsInput(c *fiber.Ctx) (cycleSettingsInput, string) {
	input := cycleSettingsInput{}
	if err := c.BodyParser(&input); err != nil {
		return cycleSettingsInput{}, "invalid settings input"
	}
	if !isValidOnboardingCycleLength(input.CycleLength) {
		return cycleSettingsInput{}, "cycle length must be between 15 and 90"
	}
	if !isValidOnboardingPeriodLength(input.PeriodLength) {
		return cycleSettingsInput{}, "period length must be between 1 and 10"
	}
	return input, ""
}

func (handler *Handler) saveCycleSettings(userID uint, input cycleSettingsInput) error {
	return handler.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]any{
		"cycle_length":     input.CycleLength,
		"period_length":    input.PeriodLength,
		"auto_period_fill": input.AutoPeriodFill,
	}).Error
}

func applyCycleSettings(user *models.User, input cycleSettingsInput) {
	user.CycleLength = input.CycleLength
	user.PeriodLength = input.PeriodLength
	user.AutoPeriodFill = input.AutoPeriodFill
}

func profileUpdateStatus(previousDisplayName string, updatedDisplayName string) string {
	status := "profile_updated"
	if strings.TrimSpace(previousDisplayName) != "" && updatedDisplayName == "" {
		status = "profile_name_cleared"
	}
	return status
}
