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

	rawPeriodStatus := strings.TrimSpace(input.PeriodStatus)
	if rawPeriodStatus == "" {
		return onboardingStep1Values{}, "period status is required"
	}
	periodStatus := normalizeOnboardingPeriodStatus(rawPeriodStatus)
	if periodStatus == "" {
		return onboardingStep1Values{}, "invalid period status"
	}

	var periodEnd *time.Time
	if periodStatus == onboardingPeriodStatusFinished {
		rawPeriodEnd := strings.TrimSpace(input.PeriodEnd)
		if rawPeriodEnd == "" {
			return onboardingStep1Values{}, "period end is required"
		}
		parsedEnd, err := parseDayParam(rawPeriodEnd, handler.location)
		if err != nil {
			return onboardingStep1Values{}, "invalid period end"
		}
		if parsedEnd.Before(parsedDay) || parsedEnd.After(today) {
			return onboardingStep1Values{}, "period end must be between start and today"
		}
		periodEnd = &parsedEnd
	}

	return onboardingStep1Values{
		Start:  parsedDay,
		Status: periodStatus,
		End:    periodEnd,
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
	if !isValidOnboardingCycleLength(input.CycleLength) {
		return onboardingStep2Input{}, "cycle length must be between 15 and 90"
	}
	if !isValidOnboardingPeriodLength(input.PeriodLength) {
		return onboardingStep2Input{}, "period length must be between 1 and 14"
	}
	if !canEstimateOvulation(input.CycleLength, input.PeriodLength) {
		return onboardingStep2Input{}, "period length is incompatible with cycle length"
	}
	return input, ""
}
