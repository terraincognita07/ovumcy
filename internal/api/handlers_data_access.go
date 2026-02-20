package api

import (
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
	"gorm.io/gorm"
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

func (handler *Handler) buildSettingsViewData(c *fiber.Ctx, user *models.User, flash FlashPayload) (fiber.Map, error) {
	messages := currentMessages(c)
	status := strings.TrimSpace(flash.SettingsSuccess)
	if status == "" {
		status = strings.TrimSpace(c.Query("success"))
		if status == "" {
			status = strings.TrimSpace(c.Query("status"))
		}
	}
	errorSource := strings.TrimSpace(flash.SettingsError)
	errorKey := ""
	if status == "" {
		if errorSource == "" {
			errorSource = strings.TrimSpace(c.Query("error"))
		}
		errorKey = authErrorTranslationKey(errorSource)
	}

	persisted := models.User{}
	if err := handler.db.Select("cycle_length", "period_length").First(&persisted, user.ID).Error; err != nil {
		return nil, err
	}

	cycleLength := persisted.CycleLength
	if !isValidOnboardingCycleLength(cycleLength) {
		cycleLength = 28
	}
	periodLength := persisted.PeriodLength
	if !isValidOnboardingPeriodLength(periodLength) {
		periodLength = 5
	}
	user.CycleLength = cycleLength
	user.PeriodLength = periodLength

	data := fiber.Map{
		"Title":        localizedPageTitle(messages, "meta.title.settings", "Lume | Settings"),
		"CurrentUser":  user,
		"ErrorKey":     errorKey,
		"SuccessKey":   settingsStatusTranslationKey(status),
		"CycleLength":  cycleLength,
		"PeriodLength": periodLength,
	}

	if user.Role == models.RoleOwner {
		totalEntries, firstDate, lastDate, err := handler.fetchExportSummary(user.ID)
		if err != nil {
			return nil, err
		}
		data["ExportTotalEntries"] = int(totalEntries)
		data["HasExportData"] = totalEntries > 0
		data["ExportDateFrom"] = firstDate
		data["ExportDateTo"] = lastDate
	}

	return data, nil
}

func (handler *Handler) fetchExportData(userID uint, from *time.Time, to *time.Time) ([]models.DailyLog, map[uint]string, error) {
	logs := make([]models.DailyLog, 0)
	query := handler.db.Where("user_id = ?", userID)
	if from != nil {
		fromKey := dateAtLocation(*from, handler.location).Format("2006-01-02")
		query = query.Where("substr(date, 1, 10) >= ?", fromKey)
	}
	if to != nil {
		toKey := dateAtLocation(*to, handler.location).Format("2006-01-02")
		query = query.Where("substr(date, 1, 10) <= ?", toKey)
	}
	if err := query.Order("date ASC").Find(&logs).Error; err != nil {
		return nil, nil, err
	}

	symptoms, err := handler.fetchSymptoms(userID)
	if err != nil {
		return nil, nil, err
	}

	symptomNames := make(map[uint]string, len(symptoms))
	for _, symptom := range symptoms {
		symptomNames[symptom.ID] = symptom.Name
	}

	return logs, symptomNames, nil
}

func (handler *Handler) fetchExportSummary(userID uint) (int64, string, string, error) {
	return handler.fetchExportSummaryForRange(userID, nil, nil)
}

func (handler *Handler) fetchExportSummaryForRange(userID uint, from *time.Time, to *time.Time) (int64, string, string, error) {
	queryWithRange := func() *gorm.DB {
		query := handler.db.Where("user_id = ?", userID)
		if from != nil {
			fromKey := dateAtLocation(*from, handler.location).Format("2006-01-02")
			query = query.Where("substr(date, 1, 10) >= ?", fromKey)
		}
		if to != nil {
			toKey := dateAtLocation(*to, handler.location).Format("2006-01-02")
			query = query.Where("substr(date, 1, 10) <= ?", toKey)
		}
		return query
	}

	var total int64
	if err := queryWithRange().Model(&models.DailyLog{}).Count(&total).Error; err != nil {
		return 0, "", "", err
	}
	if total == 0 {
		return 0, "", "", nil
	}

	var first models.DailyLog
	if err := queryWithRange().Select("date").Order("date ASC").First(&first).Error; err != nil {
		return 0, "", "", err
	}

	var last models.DailyLog
	if err := queryWithRange().Select("date").Order("date DESC").First(&last).Error; err != nil {
		return 0, "", "", err
	}

	return total,
		dateAtLocation(first.Date, handler.location).Format("2006-01-02"),
		dateAtLocation(last.Date, handler.location).Format("2006-01-02"),
		nil
}

func (handler *Handler) fetchLogsForUser(userID uint, from time.Time, to time.Time) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	fromKey := dateAtLocation(from, handler.location).Format("2006-01-02")
	toKey := dateAtLocation(to, handler.location).Format("2006-01-02")
	err := handler.db.
		Where("user_id = ? AND substr(date, 1, 10) >= ? AND substr(date, 1, 10) <= ?", userID, fromKey, toKey).
		Order("substr(date, 1, 10) ASC, id ASC").
		Find(&logs).Error
	return logs, err
}

func (handler *Handler) fetchAllLogsForUser(userID uint) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	err := handler.db.Where("user_id = ?", userID).Order("date ASC").Find(&logs).Error
	return logs, err
}

func (handler *Handler) fetchLogByDate(userID uint, day time.Time) (models.DailyLog, error) {
	entry := models.DailyLog{}
	dayStart, _ := dayRange(day, handler.location)
	dayKey := dayStart.Format("2006-01-02")
	result := handler.db.
		Where("user_id = ? AND substr(date, 1, 10) = ?", userID, dayKey).
		Order("date DESC, id DESC").
		Limit(1).
		Find(&entry)
	if result.Error != nil {
		return models.DailyLog{}, result.Error
	}
	if result.RowsAffected == 0 {
		return models.DailyLog{
			UserID:     userID,
			Date:       dayStart,
			Flow:       models.FlowNone,
			SymptomIDs: []uint{},
		}, nil
	}
	return entry, nil
}

func (handler *Handler) deleteDailyLogByDate(userID uint, day time.Time) error {
	dayKey := dateAtLocation(day, handler.location).Format("2006-01-02")
	return handler.db.Where("user_id = ? AND substr(date, 1, 10) = ?", userID, dayKey).Delete(&models.DailyLog{}).Error
}

func (handler *Handler) dayHasDataForDate(userID uint, day time.Time) (bool, error) {
	dayKey := dateAtLocation(day, handler.location).Format("2006-01-02")
	entries := make([]models.DailyLog, 0)
	if err := handler.db.
		Select("is_period", "flow", "symptom_ids", "notes").
		Where("user_id = ? AND substr(date, 1, 10) = ?", userID, dayKey).
		Find(&entries).Error; err != nil {
		return false, err
	}
	for _, entry := range entries {
		if dayHasData(entry) {
			return true, nil
		}
	}
	return false, nil
}

func (handler *Handler) refreshUserLastPeriodStart(userID uint) error {
	periodLogs := make([]models.DailyLog, 0)
	if err := handler.db.
		Select("date", "is_period").
		Where("user_id = ? AND is_period = ?", userID, true).
		Order("date ASC").
		Find(&periodLogs).Error; err != nil {
		return err
	}

	starts := services.DetectCycleStarts(periodLogs)
	if len(starts) == 0 {
		return handler.db.Model(&models.User{}).Where("id = ?", userID).Update("last_period_start", nil).Error
	}

	latest := dateAtLocation(starts[len(starts)-1], handler.location)
	return handler.db.Model(&models.User{}).Where("id = ?", userID).Update("last_period_start", latest).Error
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
