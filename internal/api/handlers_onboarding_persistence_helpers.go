package api

import "github.com/terraincognita07/ovumcy/internal/models"

func (handler *Handler) saveOnboardingStep1(user *models.User, values onboardingStep1Values) error {
	updates := map[string]any{
		"last_period_start": values.Start,
	}
	if err := handler.db.Model(&models.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}

	user.LastPeriodStart = &values.Start
	return nil
}

func (handler *Handler) saveOnboardingStep2(user *models.User, values onboardingStep2Input) error {
	values.CycleLength, values.PeriodLength = sanitizeOnboardingCycleAndPeriod(values.CycleLength, values.PeriodLength)

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
