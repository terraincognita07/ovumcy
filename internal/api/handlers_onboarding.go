package api

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
)

const (
	onboardingPeriodStatusOngoing  = "ongoing"
	onboardingPeriodStatusFinished = "finished"
)

func (handler *Handler) ShowOnboarding(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return c.Redirect("/login", fiber.StatusSeeOther)
	}
	if !requiresOnboarding(user) {
		return c.Redirect("/dashboard", fiber.StatusSeeOther)
	}

	now := dateAtLocation(time.Now().In(handler.location), handler.location)
	data := handler.buildOnboardingViewData(c, user, now)
	return handler.render(c, "onboarding", data)
}

func (handler *Handler) OnboardingStep1(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if !requiresOnboarding(user) {
		return redirectOrJSON(c, "/dashboard")
	}

	today := dateAtLocation(time.Now().In(handler.location), handler.location)
	values, validationError := handler.parseOnboardingStep1Values(c, today)
	if validationError != "" {
		return apiError(c, fiber.StatusBadRequest, validationError)
	}

	if err := handler.saveOnboardingStep1(user, values); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to save onboarding step")
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	if isHTMX(c) {
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.Redirect("/onboarding", fiber.StatusSeeOther)
}

func (handler *Handler) OnboardingStep2(c *fiber.Ctx) error {
	user, ok := currentUser(c)
	if !ok {
		return apiError(c, fiber.StatusUnauthorized, "unauthorized")
	}
	if !requiresOnboarding(user) {
		return redirectOrJSON(c, "/dashboard")
	}

	values, validationError := parseOnboardingStep2Input(c)
	if validationError != "" {
		return apiError(c, fiber.StatusBadRequest, validationError)
	}

	if err := handler.saveOnboardingStep2(user, values); err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to save onboarding step")
	}

	if acceptsJSON(c) {
		return c.JSON(fiber.Map{"ok": true})
	}
	if isHTMX(c) {
		return c.SendStatus(fiber.StatusNoContent)
	}
	return c.Redirect("/onboarding", fiber.StatusSeeOther)
}

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
