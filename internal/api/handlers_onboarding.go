package api

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
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

	data := fiber.Map{
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
		PeriodStatus    string `json:"period_status" form:"period_status"`
		PeriodEnd       string `json:"period_end" form:"period_end"`
	}{}
	if err := c.BodyParser(&input); err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid input")
	}

	rawLastPeriodStart := strings.TrimSpace(input.LastPeriodStart)
	if rawLastPeriodStart == "" {
		return apiError(c, fiber.StatusBadRequest, "date is required")
	}

	parsedDay, err := parseDayParam(rawLastPeriodStart, handler.location)
	if err != nil {
		return apiError(c, fiber.StatusBadRequest, "invalid last period start")
	}

	today := dateAtLocation(time.Now().In(handler.location), handler.location)
	minDate := today.AddDate(0, 0, -60)
	if parsedDay.After(today) || parsedDay.Before(minDate) {
		return apiError(c, fiber.StatusBadRequest, "last period start must be within last 60 days")
	}

	rawPeriodStatus := strings.TrimSpace(input.PeriodStatus)
	if rawPeriodStatus == "" {
		return apiError(c, fiber.StatusBadRequest, "period status is required")
	}
	periodStatus := normalizeOnboardingPeriodStatus(rawPeriodStatus)
	if periodStatus == "" {
		return apiError(c, fiber.StatusBadRequest, "invalid period status")
	}

	var periodEnd *time.Time
	if periodStatus == onboardingPeriodStatusFinished {
		rawPeriodEnd := strings.TrimSpace(input.PeriodEnd)
		if rawPeriodEnd == "" {
			return apiError(c, fiber.StatusBadRequest, "period end is required")
		}
		parsedEnd, err := parseDayParam(rawPeriodEnd, handler.location)
		if err != nil {
			return apiError(c, fiber.StatusBadRequest, "invalid period end")
		}
		if parsedEnd.Before(parsedDay) || parsedEnd.After(today) {
			return apiError(c, fiber.StatusBadRequest, "period end must be between start and today")
		}
		periodEnd = &parsedEnd
	}

	updates := map[string]any{
		"last_period_start":        parsedDay,
		"onboarding_period_status": periodStatus,
		"onboarding_period_end":    periodEnd,
	}
	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return apiError(c, fiber.StatusInternalServerError, "failed to save onboarding step")
	}
	user.LastPeriodStart = &parsedDay
	user.OnboardingPeriodStatus = periodStatus
	user.OnboardingPeriodEnd = periodEnd

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

	today := dateAtLocation(time.Now().In(handler.location), handler.location)
	startDay := dateAtLocation(*user.LastPeriodStart, handler.location)
	endDay := startDay
	if err := handler.db.Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.First(&current, user.ID).Error; err != nil {
			return err
		}
		if current.LastPeriodStart == nil {
			return errors.New("complete onboarding steps first")
		}
		startDay = dateAtLocation(*current.LastPeriodStart, handler.location)

		status := normalizeOnboardingPeriodStatus(current.OnboardingPeriodStatus)
		if status == "" {
			status = onboardingPeriodStatusOngoing
		}
		if status == onboardingPeriodStatusFinished {
			if current.OnboardingPeriodEnd == nil {
				return errors.New("complete onboarding steps first")
			}
			endDay = dateAtLocation(*current.OnboardingPeriodEnd, handler.location)
			if endDay.Before(startDay) || endDay.After(today) {
				return errors.New("complete onboarding steps first")
			}
		} else {
			periodLength := current.PeriodLength
			if !isValidOnboardingPeriodLength(periodLength) {
				periodLength = models.DefaultPeriodLength
			}
			endDay = startDay.AddDate(0, 0, periodLength-1)
		}

		if err := handler.upsertOnboardingPeriodRange(tx, current.ID, startDay, endDay); err != nil {
			return err
		}

		if err := tx.Model(&models.User{}).Where("id = ?", current.ID).Updates(map[string]any{
			"last_period_start":        startDay,
			"onboarding_completed":     true,
			"onboarding_period_status": "",
			"onboarding_period_end":    nil,
		}).Error; err != nil {
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
	user.OnboardingPeriodStatus = ""
	user.OnboardingPeriodEnd = nil
	return redirectOrJSON(c, "/dashboard")
}

func normalizeOnboardingPeriodStatus(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case onboardingPeriodStatusOngoing:
		return onboardingPeriodStatusOngoing
	case onboardingPeriodStatusFinished:
		return onboardingPeriodStatusFinished
	default:
		return ""
	}
}

func (handler *Handler) upsertOnboardingPeriodRange(tx *gorm.DB, userID uint, startDay time.Time, endDay time.Time) error {
	if endDay.Before(startDay) {
		return errors.New("invalid onboarding range")
	}

	for cursor := startDay; !cursor.After(endDay); cursor = cursor.AddDate(0, 0, 1) {
		day := dateAtLocation(cursor, handler.location)
		dayKey := day.Format("2006-01-02")

		var entry models.DailyLog
		result := tx.
			Where("user_id = ? AND substr(date, 1, 10) = ?", userID, dayKey).
			Order("date DESC, id DESC").
			First(&entry)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			entry = models.DailyLog{
				UserID:     userID,
				Date:       day,
				IsPeriod:   true,
				Flow:       models.FlowNone,
				SymptomIDs: []uint{},
			}
			if err := tx.Create(&entry).Error; err != nil {
				return err
			}
			continue
		}
		if result.Error != nil {
			return result.Error
		}
		if err := tx.Model(&entry).Updates(map[string]any{
			"is_period": true,
			"flow":      models.FlowNone,
		}).Error; err != nil {
			return err
		}
	}

	return nil
}
