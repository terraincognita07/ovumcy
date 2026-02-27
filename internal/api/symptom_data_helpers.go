package api

import (
	"errors"

	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) validateSymptomIDs(userID uint, ids []uint) ([]uint, error) {
	handler.ensureDependencies()
	filtered, err := handler.symptomService.ValidateSymptomIDs(userID, ids)
	if err != nil {
		if errors.Is(err, services.ErrInvalidSymptomID) {
			return nil, errors.New("invalid symptom id")
		}
		return nil, errors.New("invalid symptom id")
	}
	return filtered, nil
}
