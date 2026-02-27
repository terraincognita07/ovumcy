package services

import (
	"errors"
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type stubSymptomRepo struct {
	countByIDs int64
	countErr   error
	builtinCnt int64

	createErr  error
	findErr    error
	deleteErr  error
	findResult models.SymptomType
	created    []models.SymptomType
	deleted    []models.SymptomType
	listed     []models.SymptomType
}

func (stub *stubSymptomRepo) CountBuiltinByUser(uint) (int64, error) {
	return stub.builtinCnt, nil
}

func (stub *stubSymptomRepo) CountByUserAndIDs(uint, []uint) (int64, error) {
	return stub.countByIDs, stub.countErr
}

func (stub *stubSymptomRepo) ListByUser(uint) ([]models.SymptomType, error) {
	result := make([]models.SymptomType, len(stub.listed))
	copy(result, stub.listed)
	return result, nil
}

func (stub *stubSymptomRepo) Create(symptom *models.SymptomType) error {
	if stub.createErr != nil {
		return stub.createErr
	}
	if symptom.ID == 0 {
		symptom.ID = uint(len(stub.created) + 1)
	}
	stub.created = append(stub.created, *symptom)
	return nil
}

func (stub *stubSymptomRepo) CreateBatch([]models.SymptomType) error {
	stub.builtinCnt = 1
	return nil
}

func (stub *stubSymptomRepo) FindByIDForUser(uint, uint) (models.SymptomType, error) {
	if stub.findErr != nil {
		return models.SymptomType{}, stub.findErr
	}
	return stub.findResult, nil
}

func (stub *stubSymptomRepo) Delete(symptom *models.SymptomType) error {
	if stub.deleteErr != nil {
		return stub.deleteErr
	}
	stub.deleted = append(stub.deleted, *symptom)
	return nil
}

type stubSymptomLogRepo struct {
	listErr   error
	updateErr error
	logs      []models.DailyLog
	updated   []models.DailyLog
}

func (stub *stubSymptomLogRepo) ListByUser(uint) ([]models.DailyLog, error) {
	if stub.listErr != nil {
		return nil, stub.listErr
	}
	result := make([]models.DailyLog, len(stub.logs))
	copy(result, stub.logs)
	return result, nil
}

func (stub *stubSymptomLogRepo) UpdateSymptomIDs(entry *models.DailyLog) error {
	if stub.updateErr != nil {
		return stub.updateErr
	}
	stub.updated = append(stub.updated, *entry)
	for index := range stub.logs {
		if stub.logs[index].ID == entry.ID {
			stub.logs[index].SymptomIDs = entry.SymptomIDs
		}
	}
	return nil
}

func TestValidateSymptomIDsSortsAndDeduplicates(t *testing.T) {
	service := NewSymptomService(&stubSymptomRepo{countByIDs: 3}, &stubSymptomLogRepo{})

	ids, err := service.ValidateSymptomIDs(10, []uint{3, 1, 3, 2})
	if err != nil {
		t.Fatalf("ValidateSymptomIDs() unexpected error: %v", err)
	}
	if len(ids) != 3 || ids[0] != 1 || ids[1] != 2 || ids[2] != 3 {
		t.Fatalf("ValidateSymptomIDs() = %#v, want [1 2 3]", ids)
	}
}

func TestValidateSymptomIDsReturnsInvalidID(t *testing.T) {
	service := NewSymptomService(&stubSymptomRepo{countByIDs: 1}, &stubSymptomLogRepo{})

	_, err := service.ValidateSymptomIDs(10, []uint{3, 1})
	if !errors.Is(err, ErrInvalidSymptomID) {
		t.Fatalf("expected ErrInvalidSymptomID, got %v", err)
	}
}

func TestCreateSymptomForUserAppliesDefaultsAndTrim(t *testing.T) {
	repo := &stubSymptomRepo{}
	service := NewSymptomService(repo, &stubSymptomLogRepo{})

	symptom, err := service.CreateSymptomForUser(10, "  Custom  ", " ", " #A1B2C3 ")
	if err != nil {
		t.Fatalf("CreateSymptomForUser() unexpected error: %v", err)
	}
	if symptom.UserID != 10 {
		t.Fatalf("expected user_id 10, got %d", symptom.UserID)
	}
	if symptom.Name != "Custom" {
		t.Fatalf("expected trimmed name Custom, got %q", symptom.Name)
	}
	if symptom.Icon != defaultSymptomIcon {
		t.Fatalf("expected default icon %q, got %q", defaultSymptomIcon, symptom.Icon)
	}
	if symptom.Color != "#A1B2C3" {
		t.Fatalf("expected trimmed color #A1B2C3, got %q", symptom.Color)
	}
	if len(repo.created) != 1 {
		t.Fatalf("expected one create call, got %d", len(repo.created))
	}
}

func TestCreateSymptomForUserRejectsInvalidColor(t *testing.T) {
	service := NewSymptomService(&stubSymptomRepo{}, &stubSymptomLogRepo{})

	_, err := service.CreateSymptomForUser(10, "Custom", "A", "not-color")
	if !errors.Is(err, ErrInvalidSymptomColor) {
		t.Fatalf("expected ErrInvalidSymptomColor, got %v", err)
	}
}

func TestCreateSymptomForUserReturnsTypedCreateError(t *testing.T) {
	service := NewSymptomService(&stubSymptomRepo{createErr: errors.New("insert failed")}, &stubSymptomLogRepo{})

	_, err := service.CreateSymptomForUser(10, "Custom", "A", "#123456")
	if !errors.Is(err, ErrCreateSymptomFailed) {
		t.Fatalf("expected ErrCreateSymptomFailed, got %v", err)
	}
}

func TestDeleteSymptomForUserRejectsBuiltin(t *testing.T) {
	service := NewSymptomService(&stubSymptomRepo{
		findResult: models.SymptomType{ID: 7, UserID: 10, IsBuiltin: true},
	}, &stubSymptomLogRepo{})

	err := service.DeleteSymptomForUser(10, 7)
	if !errors.Is(err, ErrBuiltinSymptomDeleteForbidden) {
		t.Fatalf("expected ErrBuiltinSymptomDeleteForbidden, got %v", err)
	}
}

func TestDeleteSymptomForUserReturnsTypedNotFound(t *testing.T) {
	service := NewSymptomService(&stubSymptomRepo{findErr: errors.New("not found")}, &stubSymptomLogRepo{})

	err := service.DeleteSymptomForUser(10, 7)
	if !errors.Is(err, ErrSymptomNotFound) {
		t.Fatalf("expected ErrSymptomNotFound, got %v", err)
	}
}

func TestDeleteSymptomForUserDeletesAndCleansLogs(t *testing.T) {
	repo := &stubSymptomRepo{
		findResult: models.SymptomType{ID: 7, UserID: 10, IsBuiltin: false},
	}
	logs := &stubSymptomLogRepo{
		logs: []models.DailyLog{
			{ID: 1, UserID: 10, SymptomIDs: []uint{7, 8, 7}},
			{ID: 2, UserID: 10, SymptomIDs: []uint{8}},
		},
	}
	service := NewSymptomService(repo, logs)

	if err := service.DeleteSymptomForUser(10, 7); err != nil {
		t.Fatalf("DeleteSymptomForUser() unexpected error: %v", err)
	}
	if len(repo.deleted) != 1 || repo.deleted[0].ID != 7 {
		t.Fatalf("expected symptom id=7 to be deleted, got %#v", repo.deleted)
	}
	if len(logs.updated) != 1 || logs.updated[0].ID != 1 {
		t.Fatalf("expected one updated log id=1, got %#v", logs.updated)
	}
	if len(logs.logs[0].SymptomIDs) != 1 || logs.logs[0].SymptomIDs[0] != 8 {
		t.Fatalf("expected symptom cleanup in first log, got %#v", logs.logs[0].SymptomIDs)
	}
}

func TestDeleteSymptomForUserReturnsTypedCleanupError(t *testing.T) {
	repo := &stubSymptomRepo{
		findResult: models.SymptomType{ID: 7, UserID: 10, IsBuiltin: false},
	}
	logs := &stubSymptomLogRepo{
		listErr: errors.New("list failed"),
	}
	service := NewSymptomService(repo, logs)

	err := service.DeleteSymptomForUser(10, 7)
	if !errors.Is(err, ErrCleanSymptomLogsFailed) {
		t.Fatalf("expected ErrCleanSymptomLogsFailed, got %v", err)
	}
}

func TestCalculateFrequenciesIncludesTotalDaysContext(t *testing.T) {
	repo := &stubSymptomRepo{
		builtinCnt: 1,
		listed: []models.SymptomType{
			{ID: 1, Name: "A", Icon: "A"},
			{ID: 2, Name: "B", Icon: "B"},
		},
	}
	service := NewSymptomService(repo, &stubSymptomLogRepo{})

	logs := []models.DailyLog{
		{SymptomIDs: []uint{1}},
		{SymptomIDs: []uint{1, 2}},
		{SymptomIDs: []uint{}},
	}

	result, err := service.CalculateFrequencies(10, logs)
	if err != nil {
		t.Fatalf("CalculateFrequencies() unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 symptom frequencies, got %d", len(result))
	}
	if result[0].Name != "A" || result[0].Count != 2 {
		t.Fatalf("expected first result A x2, got %#v", result[0])
	}
	for _, item := range result {
		if item.TotalDays != len(logs) {
			t.Fatalf("expected total days %d, got %d", len(logs), item.TotalDays)
		}
	}
}

func TestCalculateFrequenciesSortsByCountThenName(t *testing.T) {
	repo := &stubSymptomRepo{
		builtinCnt: 1,
		listed: []models.SymptomType{
			{ID: 1, Name: "Beta", Icon: "B"},
			{ID: 2, Name: "Alpha", Icon: "A"},
		},
	}
	service := NewSymptomService(repo, &stubSymptomLogRepo{})

	logs := []models.DailyLog{
		{SymptomIDs: []uint{1}},
		{SymptomIDs: []uint{2}},
	}

	result, err := service.CalculateFrequencies(10, logs)
	if err != nil {
		t.Fatalf("CalculateFrequencies() unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 symptom frequencies, got %d", len(result))
	}
	if result[0].Name != "Alpha" || result[1].Name != "Beta" {
		t.Fatalf("expected alphabetical tie-break, got %#v", result)
	}
}
