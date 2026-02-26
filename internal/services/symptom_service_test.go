package services

import (
	"errors"
	"testing"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type stubSymptomRepo struct {
	countByIDs int64
	countErr   error
}

func (stub *stubSymptomRepo) CountBuiltinByUser(uint) (int64, error) {
	return 0, nil
}

func (stub *stubSymptomRepo) CountByUserAndIDs(uint, []uint) (int64, error) {
	return stub.countByIDs, stub.countErr
}

func (stub *stubSymptomRepo) ListByUser(uint) ([]models.SymptomType, error) {
	return nil, nil
}

func (stub *stubSymptomRepo) Create(*models.SymptomType) error {
	return nil
}

func (stub *stubSymptomRepo) CreateBatch([]models.SymptomType) error {
	return nil
}

func (stub *stubSymptomRepo) FindByIDForUser(uint, uint) (models.SymptomType, error) {
	return models.SymptomType{}, nil
}

func (stub *stubSymptomRepo) Delete(*models.SymptomType) error {
	return nil
}

type stubSymptomLogRepo struct{}

func (stub *stubSymptomLogRepo) ListByUser(uint) ([]models.DailyLog, error) {
	return nil, nil
}

func (stub *stubSymptomLogRepo) UpdateSymptomIDs(*models.DailyLog) error {
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
