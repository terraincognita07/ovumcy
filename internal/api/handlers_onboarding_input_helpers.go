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

	minDate := today.AddDate(0, 0, -60)
	if parsedDay.After(today) || parsedDay.Before(minDate) {
		return onboardingStep1Values{}, "last period start must be within last 60 days"
	}

	// Legacy compatibility: if old clients still submit period_end, treat it as
	// exclusive (first clean day) and infer period length from date difference.
	inferredPeriodLength := 0
	rawPeriodEnd := strings.TrimSpace(input.PeriodEnd)
	if rawPeriodEnd != "" {
		parsedEnd, err := parseDayParam(rawPeriodEnd, handler.location)
		if err == nil {
			days := int(parsedEnd.Sub(parsedDay).Hours() / 24)
			if days >= 1 {
				inferredPeriodLength = clampOnboardingPeriodLength(days)
			}
		}
	}

	return onboardingStep1Values{
		Start:                parsedDay,
		InferredPeriodLength: inferredPeriodLength,
	}, ""
}

func parseOnboardingStep2Input(c *fiber.Ctx, location *time.Location) (onboardingStep2Input, string) {
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
	input.PeriodLength = inferLegacyOnboardingPeriodLength(c, location, input.PeriodLength)
	_, input.PeriodLength = sanitizeOnboardingCycleAndPeriod(input.CycleLength, input.PeriodLength)

	return input, ""
}

func inferLegacyOnboardingPeriodLength(c *fiber.Ctx, location *time.Location, fallbackPeriodLength int) int {
	rawStart := strings.TrimSpace(c.FormValue("last_period_start"))
	rawEnd := strings.TrimSpace(c.FormValue("period_end"))
	if rawStart == "" || rawEnd == "" {
		return clampOnboardingPeriodLength(fallbackPeriodLength)
	}

	startDay, startErr := parseDayParam(rawStart, location)
	endDay, endErr := parseDayParam(rawEnd, location)
	if startErr != nil || endErr != nil {
		return clampOnboardingPeriodLength(fallbackPeriodLength)
	}

	days := int(endDay.Sub(startDay).Hours() / 24)
	if days < 1 {
		return clampOnboardingPeriodLength(fallbackPeriodLength)
	}
	return clampOnboardingPeriodLength(days)
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
