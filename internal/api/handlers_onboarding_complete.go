package api

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

func (handler *Handler) OnboardingComplete(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if !requiresOnboarding(user) {
		return redirectOrJSON(c, "/dashboard")
	}
	if user.LastPeriodStart == nil {
		return apiError(c, fiber.StatusBadRequest, "complete onboarding steps first")
	}

	today := dateAtLocation(time.Now().In(handler.location), handler.location)
	startDay, err := handler.completeOnboardingForUser(user.ID, today)
	if err != nil {
		if errors.Is(err, errOnboardingStepsRequired) {
			return apiError(c, fiber.StatusBadRequest, "complete onboarding steps first")
		}
		return apiError(c, fiber.StatusInternalServerError, "failed to finish onboarding")
	}

	user.OnboardingCompleted = true
	user.LastPeriodStart = &startDay
	user.OnboardingPeriodStatus = ""
	user.OnboardingPeriodEnd = nil
	return redirectOrJSON(c, "/dashboard")
}
