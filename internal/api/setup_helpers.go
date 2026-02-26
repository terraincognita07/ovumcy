package api

func (handler *Handler) requiresInitialSetup() (bool, error) {
	handler.ensureDependencies()
	return handler.setupService.RequiresInitialSetup()
}
