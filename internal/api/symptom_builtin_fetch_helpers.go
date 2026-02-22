package api

import (
	"sort"
	"strings"

	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) fetchSymptoms(userID uint) ([]models.SymptomType, error) {
	if err := handler.ensureBuiltinSymptoms(userID); err != nil {
		return nil, err
	}

	symptoms := make([]models.SymptomType, 0)
	if err := handler.db.Where("user_id = ?", userID).Find(&symptoms).Error; err != nil {
		return nil, err
	}

	sortSymptomsByBuiltinAndName(symptoms)
	return symptoms, nil
}

func sortSymptomsByBuiltinAndName(symptoms []models.SymptomType) {
	builtinOrder := builtinSymptomOrderMap()

	sort.Slice(symptoms, func(i, j int) bool {
		left := symptoms[i]
		right := symptoms[j]
		if left.IsBuiltin != right.IsBuiltin {
			return left.IsBuiltin
		}
		if left.IsBuiltin && right.IsBuiltin {
			leftIndex, leftHas := builtinOrder[strings.ToLower(strings.TrimSpace(left.Name))]
			rightIndex, rightHas := builtinOrder[strings.ToLower(strings.TrimSpace(right.Name))]
			switch {
			case leftHas && rightHas && leftIndex != rightIndex:
				return leftIndex < rightIndex
			case leftHas != rightHas:
				return leftHas
			}
		}
		return strings.ToLower(strings.TrimSpace(left.Name)) < strings.ToLower(strings.TrimSpace(right.Name))
	})
}

func builtinSymptomOrderMap() map[string]int {
	order := make(map[string]int)
	for index, symptom := range models.DefaultBuiltinSymptoms() {
		order[strings.ToLower(strings.TrimSpace(symptom.Name))] = index
	}
	return order
}
