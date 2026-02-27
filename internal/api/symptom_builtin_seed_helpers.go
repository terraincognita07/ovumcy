package api

func (handler *Handler) seedBuiltinSymptoms(userID uint) error {
	handler.ensureDependencies()
	return handler.symptomService.SeedBuiltinSymptoms(userID)
}
