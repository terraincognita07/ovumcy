package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) parseCycleSettingsInput(c *fiber.Ctx) (cycleSettingsInput, string) {
	input := cycleSettingsInput{}

	contentType := strings.ToLower(c.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") {
		if err := c.BodyParser(&input); err != nil {
			return cycleSettingsInput{}, "invalid settings input"
		}
		input.LastPeriodStart = strings.TrimSpace(input.LastPeriodStart)
		if input.LastPeriodStart != "" {
			input.LastPeriodStartSet = true
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
			CycleLength:        cycleLength,
			PeriodLength:       periodLength,
			AutoPeriodFill:     parseBoolValue(c.FormValue("auto_period_fill")),
			LastPeriodStart:    strings.TrimSpace(c.FormValue("last_period_start")),
			LastPeriodStartSet: c.Request().PostArgs().Has("last_period_start"),
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

	if input.LastPeriodStartSet && input.LastPeriodStart != "" {
		parsedDay, err := parseDayParam(input.LastPeriodStart, handler.location)
		if err != nil {
			return cycleSettingsInput{}, "invalid cycle start date"
		}
		minCycleStart, today := currentYearDateBounds(time.Now().In(handler.location), handler.location)
		if parsedDay.Before(minCycleStart) || parsedDay.After(today) {
			return cycleSettingsInput{}, "invalid cycle start date"
		}
		input.LastPeriodStart = parsedDay.Format("2006-01-02")
	}

	return input, ""
}

func (handler *Handler) saveCycleSettings(userID uint, input cycleSettingsInput) error {
	updates := map[string]any{
		"cycle_length":     input.CycleLength,
		"period_length":    input.PeriodLength,
		"auto_period_fill": input.AutoPeriodFill,
	}
	if input.LastPeriodStartSet {
		if input.LastPeriodStart == "" {
			updates["last_period_start"] = nil
		} else {
			parsedDay, err := parseDayParam(input.LastPeriodStart, handler.location)
			if err != nil {
				return err
			}
			updates["last_period_start"] = parsedDay
		}
	}
	return handler.db.Model(&models.User{}).Where("id = ?", userID).Updates(updates).Error
}

func applyCycleSettings(user *models.User, input cycleSettingsInput, location *time.Location) {
	user.CycleLength = input.CycleLength
	user.PeriodLength = input.PeriodLength
	user.AutoPeriodFill = input.AutoPeriodFill
	if !input.LastPeriodStartSet {
		return
	}
	if input.LastPeriodStart == "" {
		user.LastPeriodStart = nil
		return
	}
	parsedDay, err := parseDayParam(input.LastPeriodStart, location)
	if err != nil {
		return
	}
	user.LastPeriodStart = &parsedDay
}
