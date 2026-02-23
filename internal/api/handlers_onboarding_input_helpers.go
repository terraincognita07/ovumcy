package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
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

	minDate, maxDate := onboardingDateBounds(today, handler.location)
	if parsedDay.Before(minDate) || parsedDay.After(maxDate) {
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

	input.CycleLength = clampOnboardingCycleLength(input.CycleLength)
	input.PeriodLength = clampOnboardingPeriodLength(input.PeriodLength)
	_, input.PeriodLength = sanitizeOnboardingCycleAndPeriod(input.CycleLength, input.PeriodLength)

	return input, ""
}

func sanitizeOnboardingCycleAndPeriod(cycleLength int, periodLength int) (int, int) {
	safeCycleLength := clampOnboardingCycleLength(cycleLength)
	safePeriodLength := clampOnboardingPeriodLength(periodLength)

	if safeCycleLength-safePeriodLength < 8 {
		safePeriodLength = safeCycleLength - 8
		if safePeriodLength < 1 {
			safePeriodLength = 1
		}
	}

	return safeCycleLength, safePeriodLength
}

func clampOnboardingCycleLength(value int) int {
	if value < 15 {
		return 15
	}
	if value > 90 {
		return 90
	}
	return value
}

func clampOnboardingPeriodLength(value int) int {
	if value < 1 {
		return 1
	}
	if value > 14 {
		return 14
	}
	return value
}
