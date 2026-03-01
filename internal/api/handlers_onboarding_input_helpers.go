package api

import (
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) parseOnboardingStep1Values(c *fiber.Ctx, today time.Time) (onboardingStep1Values, string) {
	input := onboardingStep1Input{}
	if err := c.BodyParser(&input); err != nil {
		return onboardingStep1Values{}, "invalid input"
	}

	rawLastPeriodStart := strings.TrimSpace(input.LastPeriodStart)
	if rawLastPeriodStart == "" {
		return onboardingStep1Values{}, "date is required"
	}

	parsedDay, err := parseDayParam(rawLastPeriodStart, handler.location)
	if err != nil {
		return onboardingStep1Values{}, "invalid last period start"
	}

	handler.ensureDependencies()
	if err := handler.onboardingSvc.ValidateStep1StartDate(parsedDay, today, handler.location); err != nil {
		if errors.Is(err, services.ErrOnboardingStartDateOutOfRange) {
			return onboardingStep1Values{}, "last period start must be within last 60 days"
		}
		return onboardingStep1Values{}, "last period start must be within last 60 days"
	}

	return onboardingStep1Values{
		Start: parsedDay,
	}, ""
}

func parseOnboardingStep2Input(c *fiber.Ctx) (onboardingStep2Input, string) {
	input := onboardingStep2Input{}

	contentType := strings.ToLower(c.Get("Content-Type"))
	if strings.Contains(contentType, "application/json") {
		if err := c.BodyParser(&input); err != nil {
			return onboardingStep2Input{}, "invalid input"
		}
	} else {
		cycleLength, err := strconv.Atoi(strings.TrimSpace(c.FormValue("cycle_length")))
		if err != nil {
			return onboardingStep2Input{}, "invalid input"
		}
		periodLength, err := strconv.Atoi(strings.TrimSpace(c.FormValue("period_length")))
		if err != nil {
			return onboardingStep2Input{}, "invalid input"
		}
		input = onboardingStep2Input{
			CycleLength:    cycleLength,
			PeriodLength:   periodLength,
			AutoPeriodFill: parseBoolValue(c.FormValue("auto_period_fill")),
		}
	}

	input.CycleLength, input.PeriodLength = services.SanitizeOnboardingCycleAndPeriod(input.CycleLength, input.PeriodLength)

	return input, ""
}
