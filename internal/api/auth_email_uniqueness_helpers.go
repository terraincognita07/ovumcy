package api

func (handler *Handler) registrationEmailExists(normalizedEmail string) (bool, error) {
	handler.ensureDependencies()
	return handler.authService.RegistrationEmailExists(normalizedEmail)
}
