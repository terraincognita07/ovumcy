package api

import (
	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) fetchSymptoms(userID uint) ([]models.SymptomType, error) {
	handler.ensureDependencies()
	return handler.symptomService.FetchSymptoms(userID)
}

func sortSymptomsByBuiltinAndName(symptoms []models.SymptomType) {
	services.SortSymptomsByBuiltinAndName(symptoms)
}
