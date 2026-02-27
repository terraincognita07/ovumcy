package api

func (handler *Handler) refreshUserLastPeriodStart(userID uint) error {
	handler.ensureDependencies()
	return handler.dayService.RefreshUserLastPeriodStart(userID, handler.location)
}
