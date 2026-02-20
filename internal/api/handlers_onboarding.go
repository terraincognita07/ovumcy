package api

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
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
	messages := currentMessages(c)

	lastPeriodStart := ""
	if user.LastPeriodStart != nil {
		lastPeriodStart = dateAtLocation(*user.LastPeriodStart, handler.location).Format("2006-01-02")
	}

	cycleLength := user.CycleLength
	if !isValidOnboardingCycleLength(cycleLength) {
		cycleLength = 26
	}
	periodLength := user.PeriodLength
	if !isValidOnboardingPeriodLength(periodLength) {
		periodLength = 5
	}

	data := fiber.Map{
		"Title":           localizedPageTitle(messages, "meta.title.onboarding", "Lume | Onboarding"),
		"CurrentUser":     user,
		"HideNavigation":  true,
		"MinDate":         now.AddDate(0, 0, -60).Format("2006-01-02"),
		"MaxDate":         now.Format("2006-01-02"),
		"LastPeriodStart": lastPeriodStart,
		"CycleLength":     cycleLength,
		"PeriodLength":    periodLength,
		"AutoPeriodFill":  user.AutoPeriodFill,
	}
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

	input := struct {
		LastPeriodStart string `json:"last_period_start" form:"last_period_start"`
	}{}
	if err := c.BodyParser(&input); err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid input")
	}

	parsedDay, err := parseDayParam(strings.TrimSpace(input.LastPeriodStart), handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid last period start")
	}

	today := dateAtLocation(time.Now().In(handler.location), handler.location)
	minDate := today.AddDate(0, 0, -60)
	if parsedDay.After(today) || parsedDay.Before(minDate) {
		return apiError(c, fiber.StatusBadRequest, "last period start must be within last 60 days")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Update("last_period_start", parsedDay).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to save onboarding step")
	}
	user.LastPeriodStart = &parsedDay

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

	input := struct {
		CycleLength    int  `json:"cycle_length" form:"cycle_length"`
		PeriodLength   int  `json:"period_length" form:"period_length"`
		AutoPeriodFill bool `json:"auto_period_fill" form:"auto_period_fill"`
	}{}
	if err := c.BodyParser(&input); err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid input")
	}
	if !isValidOnboardingCycleLength(input.CycleLength) {
		return apiError(c, fiber.StatusBadRequest, "cycle length must be between 15 and 90")
	}
	if !isValidOnboardingPeriodLength(input.PeriodLength) {
		return apiError(c, fiber.StatusBadRequest, "period length must be between 1 and 10")
	}

	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":     input.CycleLength,
		"period_length":    input.PeriodLength,
		"auto_period_fill": input.AutoPeriodFill,
	}).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to save onboarding step")
	}
	user.CycleLength = input.CycleLength
	user.PeriodLength = input.PeriodLength
	user.AutoPeriodFill = input.AutoPeriodFill

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

	startDay := dateAtLocation(*user.LastPeriodStart, handler.location)
	if err := handler.db.Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.First(&current, user.ID).Error; err != nil {
			return err
		}
		if current.LastPeriodStart == nil {
			return errors.New("complete onboarding steps first")
		}
		startDay = dateAtLocation(*current.LastPeriodStart, handler.location)

		var entry models.DailyLog
		result := tx.Where("user_id = ? AND date = ?", current.ID, startDay).First(&entry)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			entry = models.DailyLog{
				UserID:     current.ID,
				Date:       startDay,
				IsPeriod:   true,
				Flow:       models.FlowMedium,
				SymptomIDs: []uint{},
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
		} else if result.Error != nil {
			return result.Error
		} else {
			updates := map[string]any{"is_period": true}
			if strings.TrimSpace(entry.Flow) == "" || entry.Flow == models.FlowNone {
				updates["flow"] = models.FlowMedium
			}
			if err := tx.Model(&entry).Updates(updates).Error; err != nil {
				return err
			}
		}

		if err := tx.Model(&models.User{}).Where("id = ?", current.ID).Update("onboarding_completed", true).Error; err != nil {
			return err
		}

		return nil
	}); err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "complete onboarding steps first") {
			return apiError(c, fiber.StatusBadRequest, "complete onboarding steps first")
		}
		return apiError(c, fiber.StatusInternalServerError, "failed to finish onboarding")
	}

	user.OnboardingCompleted = true
	user.LastPeriodStart = &startDay
	return redirectOrJSON(c, "/dashboard")
}
