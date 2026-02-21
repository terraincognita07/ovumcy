package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

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
