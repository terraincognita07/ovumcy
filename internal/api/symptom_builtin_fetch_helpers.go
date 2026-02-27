package api

import (
	"github.com/terraincognita07/ovumcy/internal/models"
)

func (handler *Handler) fetchSymptoms(userID uint) ([]models.SymptomType, error) {
	handler.ensureDependencies()
	return handler.symptomService.FetchSymptoms(userID)
}
