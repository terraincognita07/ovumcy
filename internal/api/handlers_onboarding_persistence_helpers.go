package api

import "github.com/terraincognita07/ovumcy/internal/models"

func (handler *Handler) saveOnboardingStep1(user *models.User, values onboardingStep1Values) error {
	handler.ensureDependencies()
	if err := handler.onboardingSvc.SaveStep1(user.ID, values.Start); err != nil {
		return err
	}

	user.LastPeriodStart = &values.Start
	return nil
}

func (handler *Handler) saveOnboardingStep2(user *models.User, values onboardingStep2Input) error {
	handler.ensureDependencies()
	cycleLength, periodLength, err := handler.onboardingSvc.SaveStep2(user.ID, values.CycleLength, values.PeriodLength, values.AutoPeriodFill)
	if err != nil {
		return err
	}

	values.CycleLength = cycleLength
	values.PeriodLength = periodLength
	user.CycleLength = cycleLength
	user.PeriodLength = periodLength
	user.AutoPeriodFill = values.AutoPeriodFill
	return nil
}
