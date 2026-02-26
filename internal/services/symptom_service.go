package services

import (
	"errors"
	"sort"
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
)

var ErrInvalidSymptomID = errors.New("invalid symptom id")

type SymptomRepository interface {
	CountBuiltinByUser(userID uint) (int64, error)
	CountByUserAndIDs(userID uint, ids []uint) (int64, error)
	ListByUser(userID uint) ([]models.SymptomType, error)
	Create(symptom *models.SymptomType) error
	CreateBatch(symptoms []models.SymptomType) error
	FindByIDForUser(symptomID uint, userID uint) (models.SymptomType, error)
	Delete(symptom *models.SymptomType) error
}

type SymptomLogRepository interface {
	ListByUser(userID uint) ([]models.DailyLog, error)
	UpdateSymptomIDs(entry *models.DailyLog) error
}

type SymptomService struct {
	symptoms SymptomRepository
	logs     SymptomLogRepository
}

func NewSymptomService(symptoms SymptomRepository, logs SymptomLogRepository) *SymptomService {
	return &SymptomService{
		symptoms: symptoms,
		logs:     logs,
	}
}

func (service *SymptomService) CreateUserSymptom(symptom *models.SymptomType) error {
	return service.symptoms.Create(symptom)
}

func (service *SymptomService) FindSymptomForUser(symptomID uint, userID uint) (models.SymptomType, error) {
	return service.symptoms.FindByIDForUser(symptomID, userID)
}

func (service *SymptomService) DeleteSymptom(symptom *models.SymptomType) error {
	return service.symptoms.Delete(symptom)
}

func (service *SymptomService) SeedBuiltinSymptoms(userID uint) error {
	count, err := service.symptoms.CountBuiltinByUser(userID)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil
	}
	return service.symptoms.CreateBatch(BuiltinSymptomRecordsForUser(userID))
}

func (service *SymptomService) EnsureBuiltinSymptoms(userID uint) error {
	existing, err := service.symptoms.ListByUser(userID)
	if err != nil {
		return err
	}
	existingByName := make(map[string]struct{}, len(existing))
	for _, symptom := range existing {
		key := strings.ToLower(strings.TrimSpace(symptom.Name))
		if key != "" {
			existingByName[key] = struct{}{}
		}
	}

	missing := MissingBuiltinSymptomsForUser(userID, existingByName)
	if len(missing) == 0 {
		return nil
	}
	return service.symptoms.CreateBatch(missing)
}

func (service *SymptomService) FetchSymptoms(userID uint) ([]models.SymptomType, error) {
	if err := service.EnsureBuiltinSymptoms(userID); err != nil {
		return nil, err
	}
	symptoms, err := service.symptoms.ListByUser(userID)
	if err != nil {
		return nil, err
	}
	SortSymptomsByBuiltinAndName(symptoms)
	return symptoms, nil
}

func (service *SymptomService) ValidateSymptomIDs(userID uint, ids []uint) ([]uint, error) {
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

	matched, err := service.symptoms.CountByUserAndIDs(userID, filtered)
	if err != nil {
		return nil, err
	}
	if int(matched) != len(filtered) {
		return nil, ErrInvalidSymptomID
	}
	sort.Slice(filtered, func(i, j int) bool { return filtered[i] < filtered[j] })
	return filtered, nil
}

func (service *SymptomService) RemoveSymptomFromLogs(userID uint, symptomID uint) error {
	logs, err := service.logs.ListByUser(userID)
	if err != nil {
		return err
	}

	for index := range logs {
		updated := RemoveUint(logs[index].SymptomIDs, symptomID)
		if len(updated) == len(logs[index].SymptomIDs) {
			continue
		}
		logs[index].SymptomIDs = updated
		if err := service.logs.UpdateSymptomIDs(&logs[index]); err != nil {
			return err
		}
	}
	return nil
}

func BuiltinSymptomRecordsForUser(userID uint) []models.SymptomType {
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
	return records
}

func MissingBuiltinSymptomsForUser(userID uint, existingByName map[string]struct{}) []models.SymptomType {
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
	return missing
}

func SortSymptomsByBuiltinAndName(symptoms []models.SymptomType) {
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
