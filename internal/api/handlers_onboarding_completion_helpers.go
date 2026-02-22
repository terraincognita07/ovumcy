package api

import (
	"errors"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

var errOnboardingStepsRequired = errors.New("complete onboarding steps first")

func (handler *Handler) completeOnboardingForUser(userID uint, today time.Time) (time.Time, error) {
	var startDay time.Time
	err := handler.db.Transaction(func(tx *gorm.DB) error {
		var current models.User
		if err := tx.First(&current, userID).Error; err != nil {
			return err
		}
		if current.LastPeriodStart == nil {
			return errOnboardingStepsRequired
		}
		startDay = dateAtLocation(*current.LastPeriodStart, handler.location)

		var endDay time.Time
		status := normalizeOnboardingPeriodStatus(current.OnboardingPeriodStatus)
		if status == "" {
			status = onboardingPeriodStatusOngoing
		}
		if status == onboardingPeriodStatusFinished {
			if current.OnboardingPeriodEnd == nil {
				return errOnboardingStepsRequired
			}
			endDay = dateAtLocation(*current.OnboardingPeriodEnd, handler.location)
			if endDay.Before(startDay) || endDay.After(today) {
				return errOnboardingStepsRequired
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

		return tx.Model(&models.User{}).Where("id = ?", current.ID).Updates(map[string]any{
			"last_period_start":        startDay,
			"onboarding_completed":     true,
			"onboarding_period_status": "",
			"onboarding_period_end":    nil,
		}).Error
	})
	if err != nil {
		return time.Time{}, err
	}
	return startDay, nil
}
