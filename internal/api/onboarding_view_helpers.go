package api

import (
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) buildOnboardingViewData(c *fiber.Ctx, user *models.User, now time.Time) fiber.Map {
	messages := currentMessages(c)

	lastPeriodStart := ""
	if user.LastPeriodStart != nil {
		lastPeriodStart = dateAtLocation(*user.LastPeriodStart, handler.location).Format("2006-01-02")
	}

	cycleLength := user.CycleLength
	if !isValidOnboardingCycleLength(cycleLength) {
		cycleLength = models.DefaultCycleLength
	}
	periodLength := user.PeriodLength
	if !isValidOnboardingPeriodLength(periodLength) {
		periodLength = models.DefaultPeriodLength
	}

	return fiber.Map{
		"Title":           localizedPageTitle(messages, "meta.title.onboarding", "Ovumcy | Onboarding"),
		"CurrentUser":     user,
		"HideNavigation":  true,
		"OnboardingStep":  parseOnboardingStep(c.Query("step")),
		"MinDate":         now.AddDate(0, 0, -60).Format("2006-01-02"),
		"MaxDate":         now.Format("2006-01-02"),
		"LastPeriodStart": lastPeriodStart,
		"CycleLength":     cycleLength,
		"PeriodLength":    periodLength,
		"AutoPeriodFill":  user.AutoPeriodFill,
	}
}

func parseOnboardingStep(raw string) int {
	step, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil {
		return 0
	}
	if step < 0 {
		return 0
	}
	if step > 3 {
		return 3
	}
	return step
}
