package api

import (
	"errors"
	"sort"
	"strings"

	"github.com/terraincognita07/lume/internal/models"
)

func (handler *Handler) seedBuiltinSymptoms(userID uint) error {
	var count int64
	if err := handler.db.Model(&models.SymptomType{}).
		Where("user_id = ? AND is_builtin = ?", userID, true).
		Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	builtin := models.DefaultBuiltinSymptoms()
	records := make([]models.SymptomType, 0, len(builtin))
	for _, symptom := range builtin {
		records = append(records, models.SymptomType{
			UserID:    userID,
			Name:      symptom.Name,
			Icon:      symptom.Icon,
			Color:     symptom.Color,
			IsBuiltin: true,
		})
	}

	return handler.db.Create(&records).Error
}

func (handler *Handler) fetchSymptoms(userID uint) ([]models.SymptomType, error) {
	if err := handler.ensureBuiltinSymptoms(userID); err != nil {
		return nil, err
	}

	symptoms := make([]models.SymptomType, 0)
	if err := handler.db.Where("user_id = ?", userID).Find(&symptoms).Error; err != nil {
		return nil, err
	}
	for index := range symptoms {
		symptoms[index].Name = normalizeLegacySymptomName(symptoms[index].Name)
	}

	builtinOrder := make(map[string]int)
	for index, symptom := range models.DefaultBuiltinSymptoms() {
		builtinOrder[strings.ToLower(strings.TrimSpace(symptom.Name))] = index
	}

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

	return symptoms, nil
}

func (handler *Handler) ensureBuiltinSymptoms(userID uint) error {
	if err := handler.db.
		Model(&models.SymptomType{}).
		Where("user_id = ? AND lower(trim(name)) = ?", userID, "fatique").
		Update("name", "Fatigue").Error; err != nil {
		return err
	}

	existing := make([]models.SymptomType, 0)
	if err := handler.db.Where("user_id = ?", userID).Find(&existing).Error; err != nil {
		return err
	}

	existingByName := make(map[string]struct{}, len(existing))
	for _, symptom := range existing {
		key := strings.ToLower(strings.TrimSpace(symptom.Name))
		if key != "" {
			existingByName[key] = struct{}{}
		}
	}

	missing := make([]models.SymptomType, 0)
	for _, symptom := range models.DefaultBuiltinSymptoms() {
		key := strings.ToLower(strings.TrimSpace(symptom.Name))
		if _, ok := existingByName[key]; ok {
			continue
		}
		missing = append(missing, models.SymptomType{
			UserID:    userID,
			Name:      symptom.Name,
			Icon:      symptom.Icon,
			Color:     symptom.Color,
			IsBuiltin: true,
		})
	}

	if len(missing) == 0 {
		return nil
	}
	return handler.db.Create(&missing).Error
}

func (handler *Handler) validateSymptomIDs(userID uint, ids []uint) ([]uint, error) {
	if len(ids) == 0 {
		return []uint{}, nil
	}

	unique := make(map[uint]struct{}, len(ids))
	for _, id := range ids {
		unique[id] = struct{}{}
	}
	filtered := make([]uint, 0, len(unique))
	for id := range unique {
		filtered = append(filtered, id)
	}

	var matched int64
	if err := handler.db.Model(&models.SymptomType{}).
		Where("user_id = ? AND id IN ?", userID, filtered).
		Count(&matched).Error; err != nil {
		return nil, err
	}
	if int(matched) != len(filtered) {
		return nil, errors.New("invalid symptom id")
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i] < filtered[j] })
	return filtered, nil
}

func (handler *Handler) removeSymptomFromLogs(userID uint, symptomID uint) error {
	logs := make([]models.DailyLog, 0)
	if err := handler.db.Where("user_id = ?", userID).Find(&logs).Error; err != nil {
		return err
	}

	for index := range logs {
		updated := removeUint(logs[index].SymptomIDs, symptomID)
		if len(updated) == len(logs[index].SymptomIDs) {
			continue
		}
		logs[index].SymptomIDs = updated
		if err := handler.db.Model(&logs[index]).
			Select("symptom_ids").
			Updates(&logs[index]).Error; err != nil {
			return err
		}
	}
	return nil
}
