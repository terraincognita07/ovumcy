package api

import (
	"errors"
	"sort"

	"github.com/terraincognita07/ovumcy/internal/models"
)

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
