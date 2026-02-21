package api

import (
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
	if err := c.BodyParser(&input); err != nil {
		return onboardingStep2Input{}, "invalid input"
	}
	if !isValidOnboardingCycleLength(input.CycleLength) {
		return onboardingStep2Input{}, "cycle length must be between 15 and 90"
	}
	if !isValidOnboardingPeriodLength(input.PeriodLength) {
		return onboardingStep2Input{}, "period length must be between 1 and 10"
	}
	return input, ""
}
