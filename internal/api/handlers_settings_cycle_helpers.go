package api

import (
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func parseCycleSettingsInput(c *fiber.Ctx) (cycleSettingsInput, string) {
	input := cycleSettingsInput{}

	contentType := strings.ToLower(c.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") {
		if err := c.BodyParser(&input); err != nil {
			return cycleSettingsInput{}, "invalid settings input"
		}
	} else {
		cycleLength, err := strconv.Atoi(strings.TrimSpace(c.FormValue("cycle_length")))
		if err != nil {
			return cycleSettingsInput{}, "invalid settings input"
		}
		periodLength, err := strconv.Atoi(strings.TrimSpace(c.FormValue("period_length")))
		if err != nil {
			return cycleSettingsInput{}, "invalid settings input"
		}
		input = cycleSettingsInput{
			CycleLength:    cycleLength,
			PeriodLength:   periodLength,
			AutoPeriodFill: parseBoolValue(c.FormValue("auto_period_fill")),
		}
	}

	if !isValidOnboardingCycleLength(input.CycleLength) {
		return cycleSettingsInput{}, "cycle length must be between 15 and 90"
	}
	if !isValidOnboardingPeriodLength(input.PeriodLength) {
		return cycleSettingsInput{}, "period length must be between 1 and 14"
	}
	if !canEstimateOvulation(input.CycleLength, input.PeriodLength) {
		return cycleSettingsInput{}, "period length is incompatible with cycle length"
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
