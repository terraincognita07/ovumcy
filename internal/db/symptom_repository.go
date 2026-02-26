package db

import (
	"github.com/terraincognita07/ovumcy/internal/models"
	"gorm.io/gorm"
)

type SymptomRepository struct {
	database *gorm.DB
}

func NewSymptomRepository(database *gorm.DB) *SymptomRepository {
	return &SymptomRepository{database: database}
}

func (repo *SymptomRepository) CountBuiltinByUser(userID uint) (int64, error) {
	var count int64
	if err := repo.database.Model(&models.SymptomType{}).
		Where("user_id = ? AND is_builtin = ?", userID, true).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *SymptomRepository) CountByUserAndIDs(userID uint, ids []uint) (int64, error) {
	var count int64
	if err := repo.database.Model(&models.SymptomType{}).
		Where("user_id = ? AND id IN ?", userID, ids).
		Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

func (repo *SymptomRepository) ListByUser(userID uint) ([]models.SymptomType, error) {
	symptoms := make([]models.SymptomType, 0)
	if err := repo.database.Where("user_id = ?", userID).Find(&symptoms).Error; err != nil {
		return nil, err
	}
	return symptoms, nil
}

func (repo *SymptomRepository) Create(symptom *models.SymptomType) error {
	return repo.database.Create(symptom).Error
}

func (repo *SymptomRepository) CreateBatch(symptoms []models.SymptomType) error {
	if len(symptoms) == 0 {
		return nil
	}
	return repo.database.Create(&symptoms).Error
}

func (repo *SymptomRepository) FindByIDForUser(symptomID uint, userID uint) (models.SymptomType, error) {
	symptom := models.SymptomType{}
	if err := repo.database.Where("id = ? AND user_id = ?", symptomID, userID).First(&symptom).Error; err != nil {
		return models.SymptomType{}, err
	}
	return symptom, nil
}

func (repo *SymptomRepository) Delete(symptom *models.SymptomType) error {
	return repo.database.Delete(symptom).Error
}
