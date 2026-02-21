package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) buildOnboardingViewData(c *fiber.Ctx, user *models.User, now time.Time) fiber.Map {
	messages := currentMessages(c)

	lastPeriodStart := ""
	if user.LastPeriodStart != nil {
		lastPeriodStart = dateAtLocation(*user.LastPeriodStart, handler.location).Format("2006-01-02")
	}
	periodEnd := ""
	if user.OnboardingPeriodEnd != nil {
		periodEnd = dateAtLocation(*user.OnboardingPeriodEnd, handler.location).Format("2006-01-02")
	}
	periodStatus := normalizeOnboardingPeriodStatus(user.OnboardingPeriodStatus)

	cycleLength := user.CycleLength
	if !isValidOnboardingCycleLength(cycleLength) {
		cycleLength = models.DefaultCycleLength
	}
	periodLength := user.PeriodLength
	if !isValidOnboardingPeriodLength(periodLength) {
		periodLength = models.DefaultPeriodLength
	}

	return fiber.Map{
		"Title":                  localizedPageTitle(messages, "meta.title.onboarding", "Lume | Onboarding"),
		"CurrentUser":            user,
		"HideNavigation":         true,
		"MinDate":                now.AddDate(0, 0, -60).Format("2006-01-02"),
		"MaxDate":                now.Format("2006-01-02"),
		"LastPeriodStart":        lastPeriodStart,
		"OnboardingPeriodStatus": periodStatus,
		"OnboardingPeriodEnd":    periodEnd,
		"CycleLength":            cycleLength,
		"PeriodLength":           periodLength,
		"AutoPeriodFill":         user.AutoPeriodFill,
	}
}
