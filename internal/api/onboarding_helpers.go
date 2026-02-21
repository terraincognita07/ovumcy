package api

import (
	"errors"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
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
		dayKey := dayStorageKey(day, handler.location)
		nextDayKey := nextDayStorageKey(day, handler.location)

		var entry models.DailyLog
		result := tx.
			Where("user_id = ? AND date >= ? AND date < ?", userID, dayKey, nextDayKey).
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
