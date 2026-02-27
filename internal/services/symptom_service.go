package services

import (
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
)

var ErrInvalidSymptomID = errors.New("invalid symptom id")

var (
	ErrInvalidSymptomName            = errors.New("invalid symptom name")
	ErrInvalidSymptomColor           = errors.New("invalid symptom color")
	ErrSymptomNotFound               = errors.New("symptom not found")
	ErrBuiltinSymptomDeleteForbidden = errors.New("built-in symptom cannot be deleted")
	ErrCreateSymptomFailed           = errors.New("create symptom failed")
	ErrDeleteSymptomFailed           = errors.New("delete symptom failed")
	ErrCleanSymptomLogsFailed        = errors.New("clean symptom logs failed")
)

const (
	maxSymptomNameLength = 80
	defaultSymptomIcon   = "✨"
)

var hexSymptomColorPattern = regexp.MustCompile(`^#[0-9A-Fa-f]{6}$`)

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

type SymptomFrequency struct {
	Name      string
	Icon      string
	Count     int
	TotalDays int
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

func (service *SymptomService) CreateSymptomForUser(userID uint, name string, icon string, color string) (models.SymptomType, error) {
	name = strings.TrimSpace(name)
	icon = strings.TrimSpace(icon)
	color = strings.TrimSpace(color)

	if name == "" || len(name) > maxSymptomNameLength {
		return models.SymptomType{}, ErrInvalidSymptomName
	}
	if icon == "" {
		icon = defaultSymptomIcon
	}
	if !hexSymptomColorPattern.MatchString(color) {
		return models.SymptomType{}, ErrInvalidSymptomColor
	}

	symptom := models.SymptomType{
		UserID:    userID,
		Name:      name,
		Icon:      icon,
		Color:     color,
		IsBuiltin: false,
	}
	if err := service.symptoms.Create(&symptom); err != nil {
		return models.SymptomType{}, fmt.Errorf("%w: %v", ErrCreateSymptomFailed, err)
	}
	return symptom, nil
}

func (service *SymptomService) FindSymptomForUser(symptomID uint, userID uint) (models.SymptomType, error) {
	return service.symptoms.FindByIDForUser(symptomID, userID)
}

func (service *SymptomService) DeleteSymptom(symptom *models.SymptomType) error {
	return service.symptoms.Delete(symptom)
}

func (service *SymptomService) DeleteSymptomForUser(userID uint, symptomID uint) error {
	symptom, err := service.symptoms.FindByIDForUser(symptomID, userID)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrSymptomNotFound, err)
	}
	if symptom.IsBuiltin {
		return ErrBuiltinSymptomDeleteForbidden
	}

	if err := service.symptoms.Delete(&symptom); err != nil {
		return fmt.Errorf("%w: %v", ErrDeleteSymptomFailed, err)
	}

	if err := service.RemoveSymptomFromLogs(userID, symptom.ID); err != nil {
		return fmt.Errorf("%w: %v", ErrCleanSymptomLogsFailed, err)
	}
	return nil
}

func (service *SymptomService) CalculateFrequencies(userID uint, logs []models.DailyLog) ([]SymptomFrequency, error) {
	if len(logs) == 0 {
		return []SymptomFrequency{}, nil
	}
	totalDays := len(logs)

	counts := make(map[uint]int)
	for _, logEntry := range logs {
		for _, id := range logEntry.SymptomIDs {
			counts[id]++
		}
	}
	if len(counts) == 0 {
		return []SymptomFrequency{}, nil
	}

	symptoms, err := service.FetchSymptoms(userID)
	if err != nil {
		return nil, err
	}

	symptomByID := make(map[uint]models.SymptomType, len(symptoms))
	for _, symptom := range symptoms {
		symptomByID[symptom.ID] = symptom
	}

	result := make([]SymptomFrequency, 0, len(counts))
	for id, count := range counts {
		if symptom, ok := symptomByID[id]; ok {
			result = append(result, SymptomFrequency{
				Name:      symptom.Name,
				Icon:      symptom.Icon,
				Count:     count,
				TotalDays: totalDays,
			})
		}
	}

	sort.Slice(result, func(i, j int) bool {
		if result[i].Count == result[j].Count {
			return result[i].Name < result[j].Name
		}
		return result[i].Count > result[j].Count
	})

	return result, nil
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
