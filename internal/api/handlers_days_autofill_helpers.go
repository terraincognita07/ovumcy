package api

import "time"

func (handler *Handler) autoFillFollowingPeriodDays(userID uint, startDay time.Time, periodLength int, flow string) error {
	handler.ensureDependencies()
	return handler.dayService.AutoFillFollowingPeriodDays(userID, startDay, periodLength, flow, handler.location)
}

func (handler *Handler) loadDayAutoFillSettings(userID uint) (int, bool, error) {
	handler.ensureDependencies()
	return handler.dayService.LoadAutoFillSettings(userID)
}

func (handler *Handler) shouldAutoFillPeriodDays(userID uint, dayStart time.Time, wasPeriod bool, autoPeriodFillEnabled bool, periodLength int) (bool, error) {
	handler.ensureDependencies()
	return handler.dayService.ShouldAutoFillPeriodDays(userID, dayStart, wasPeriod, autoPeriodFillEnabled, periodLength, handler.location)
}
