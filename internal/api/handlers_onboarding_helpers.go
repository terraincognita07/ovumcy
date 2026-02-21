package api

import (
	"errors"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

var errOnboardingStepsRequired = errors.New("complete onboarding steps first")

type onboardingStep1Input struct {
	LastPeriodStart string `json:"last_period_start" form:"last_period_start"`
	PeriodStatus    string `json:"period_status" form:"period_status"`
	PeriodEnd       string `json:"period_end" form:"period_end"`
}

type onboardingStep1Values struct {
	Start  time.Time
	Status string
	End    *time.Time
}

type onboardingStep2Input struct {
	CycleLength    int  `json:"cycle_length" form:"cycle_length"`
	PeriodLength   int  `json:"period_length" form:"period_length"`
	AutoPeriodFill bool `json:"auto_period_fill" form:"auto_period_fill"`
}

func (handler *Handler) saveOnboardingStep1(user *models.User, values onboardingStep1Values) error {
	updates := map[string]any{
		"last_period_start":        values.Start,
		"onboarding_period_status": values.Status,
		"onboarding_period_end":    values.End,
	}
	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}

	user.LastPeriodStart = &values.Start
	user.OnboardingPeriodStatus = values.Status
	user.OnboardingPeriodEnd = values.End
	return nil
}

func (handler *Handler) saveOnboardingStep2(user *models.User, values onboardingStep2Input) error {
	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(map[string]any{
		"cycle_length":     values.CycleLength,
		"period_length":    values.PeriodLength,
		"auto_period_fill": values.AutoPeriodFill,
	}).Error; err != nil {
		return err
	}

	user.CycleLength = values.CycleLength
	user.PeriodLength = values.PeriodLength
	user.AutoPeriodFill = values.AutoPeriodFill
	return nil
}

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

		endDay := startDay
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
