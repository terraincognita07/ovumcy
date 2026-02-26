package services

import (
	"errors"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

var (
	ErrDayEntryLoadFailed   = errors.New("load day entry failed")
	ErrDayEntryCreateFailed = errors.New("create day entry failed")
	ErrDayEntryUpdateFailed = errors.New("update day entry failed")
	ErrDeleteDayFailed      = errors.New("delete day failed")
	ErrSyncLastPeriodFailed = errors.New("sync last period failed")
)

type DayEntryInput struct {
	IsPeriod   bool
	Flow       string
	Notes      string
	SymptomIDs []uint
}

type DayLogRepository interface {
	ListByUser(userID uint) ([]models.DailyLog, error)
	ListByUserRange(userID uint, fromStart *time.Time, toEnd *time.Time) ([]models.DailyLog, error)
	ListByUserDayRange(userID uint, dayStart time.Time, dayEnd time.Time) ([]models.DailyLog, error)
	ListPeriodDays(userID uint) ([]models.DailyLog, error)
	FindByUserAndDayRange(userID uint, dayStart time.Time, dayEnd time.Time) (models.DailyLog, bool, error)
	Create(entry *models.DailyLog) error
	Save(entry *models.DailyLog) error
	DeleteByUserAndDayRange(userID uint, dayStart time.Time, dayEnd time.Time) error
}

type DayUserRepository interface {
	LoadSettingsByID(userID uint) (models.User, error)
	UpdateByID(userID uint, updates map[string]any) error
}

type DayService struct {
	logs  DayLogRepository
	users DayUserRepository
}

func NewDayService(logs DayLogRepository, users DayUserRepository) *DayService {
	return &DayService{
		logs:  logs,
		users: users,
	}
}

func (service *DayService) FetchLogsForUser(userID uint, from time.Time, to time.Time, location *time.Location) ([]models.DailyLog, error) {
	fromStart, _ := DayRange(from, location)
	_, toEnd := DayRange(to, location)
	return service.logs.ListByUserRange(userID, &fromStart, &toEnd)
}

func (service *DayService) FetchLogsForOptionalRange(userID uint, from *time.Time, to *time.Time, location *time.Location) ([]models.DailyLog, error) {
	var fromStart *time.Time
	var toEnd *time.Time
	if from != nil {
		start, _ := DayRange(*from, location)
		fromStart = &start
	}
	if to != nil {
		_, end := DayRange(*to, location)
		toEnd = &end
	}
	return service.logs.ListByUserRange(userID, fromStart, toEnd)
}

func (service *DayService) FetchAllLogsForUser(userID uint) ([]models.DailyLog, error) {
	return service.logs.ListByUser(userID)
}

func (service *DayService) FetchLogByDate(userID uint, day time.Time, location *time.Location) (models.DailyLog, error) {
	dayStart, dayEnd := DayRange(day, location)
	entry, found, err := service.logs.FindByUserAndDayRange(userID, dayStart, dayEnd)
	if err != nil {
		return models.DailyLog{}, err
	}
	if !found {
		return models.DailyLog{
			UserID:     userID,
			Date:       dayStart,
			Flow:       models.FlowNone,
			SymptomIDs: []uint{},
		}, nil
	}
	return entry, nil
}

func (service *DayService) DayHasDataForDate(userID uint, day time.Time, location *time.Location) (bool, error) {
	dayStart, dayEnd := DayRange(day, location)
	entries, err := service.logs.ListByUserDayRange(userID, dayStart, dayEnd)
	if err != nil {
		return false, err
	}
	for _, entry := range entries {
		if DayHasData(entry) {
			return true, nil
		}
	}
	return false, nil
}

func (service *DayService) UpsertDayEntry(userID uint, dayStart time.Time, payload DayEntryInput, location *time.Location) (models.DailyLog, bool, error) {
	dayRangeStart, dayRangeEnd := DayRange(dayStart, location)
	entry, found, err := service.logs.FindByUserAndDayRange(userID, dayRangeStart, dayRangeEnd)
	if err != nil {
		return models.DailyLog{}, false, ErrDayEntryLoadFailed
	}

	wasPeriod := false
	if found {
		wasPeriod = entry.IsPeriod
		entry.IsPeriod = payload.IsPeriod
		entry.Flow = payload.Flow
		entry.SymptomIDs = payload.SymptomIDs
		entry.Notes = payload.Notes
		if err := service.logs.Save(&entry); err != nil {
			return models.DailyLog{}, false, ErrDayEntryUpdateFailed
		}
		return entry, wasPeriod, nil
	}

	entry = models.DailyLog{
		UserID:     userID,
		Date:       dayStart,
		IsPeriod:   payload.IsPeriod,
		Flow:       payload.Flow,
		Notes:      payload.Notes,
		SymptomIDs: payload.SymptomIDs,
	}
	if err := service.logs.Create(&entry); err != nil {
		return models.DailyLog{}, false, ErrDayEntryCreateFailed
	}
	return entry, false, nil
}

func (service *DayService) DeleteDayAndRefreshLastPeriod(userID uint, day time.Time, location *time.Location) error {
	if err := service.DeleteDailyLogByDate(userID, day, location); err != nil {
		return ErrDeleteDayFailed
	}
	if err := service.RefreshUserLastPeriodStart(userID, location); err != nil {
		return ErrSyncLastPeriodFailed
	}
	return nil
}

func (service *DayService) DeleteDailyLogByDate(userID uint, day time.Time, location *time.Location) error {
	dayStart, dayEnd := DayRange(day, location)
	return service.logs.DeleteByUserAndDayRange(userID, dayStart, dayEnd)
}

func (service *DayService) RefreshUserLastPeriodStart(userID uint, location *time.Location) error {
	periodLogs, err := service.logs.ListPeriodDays(userID)
	if err != nil {
		return err
	}
	starts := DetectCycleStarts(periodLogs)
	if len(starts) == 0 {
		return service.users.UpdateByID(userID, map[string]any{"last_period_start": nil})
	}

	latest := DateAtLocation(starts[len(starts)-1], location)
	return service.users.UpdateByID(userID, map[string]any{"last_period_start": latest})
}

func (service *DayService) LoadAutoFillSettings(userID uint) (int, bool, error) {
	persisted, err := service.users.LoadSettingsByID(userID)
	if err != nil {
		return models.DefaultPeriodLength, false, err
	}
	periodLength := persisted.PeriodLength
	if periodLength < 1 || periodLength > 14 {
		periodLength = models.DefaultPeriodLength
	}
	return periodLength, persisted.AutoPeriodFill, nil
}

func (service *DayService) ShouldAutoFillPeriodDays(userID uint, dayStart time.Time, wasPeriod bool, autoPeriodFillEnabled bool, periodLength int, location *time.Location) (bool, error) {
	if !autoPeriodFillEnabled || periodLength <= 1 || wasPeriod {
		return false, nil
	}

	previousDay := dayStart.AddDate(0, 0, -1)
	previousEntry, err := service.FetchLogByDate(userID, previousDay, location)
	if err != nil {
		return false, err
	}
	hasRecentPeriod, err := service.hasPeriodInRecentDays(userID, dayStart, 3, location)
	if err != nil {
		return false, err
	}
	return !previousEntry.IsPeriod && !hasRecentPeriod, nil
}

func (service *DayService) AutoFillFollowingPeriodDays(userID uint, startDay time.Time, periodLength int, flow string, location *time.Location) error {
	if periodLength <= 1 {
		return nil
	}

	for offset := 1; offset < periodLength; offset++ {
		targetDay := DateAtLocation(startDay.AddDate(0, 0, offset), location)
		entry, err := service.FetchLogByDate(userID, targetDay, location)
		if err != nil {
			return err
		}

		if entry.ID != 0 {
			if DayHasData(entry) && !entry.IsPeriod {
				break
			}
			if entry.IsPeriod {
				continue
			}

			entry.IsPeriod = true
			entry.Flow = flow
			if err := service.logs.Save(&entry); err != nil {
				return err
			}
			continue
		}

		newEntry := models.DailyLog{
			UserID:     userID,
			Date:       targetDay,
			IsPeriod:   true,
			Flow:       flow,
			SymptomIDs: []uint{},
		}
		if err := service.logs.Create(&newEntry); err != nil {
			return err
		}
	}

	return nil
}

func (service *DayService) hasPeriodInRecentDays(userID uint, day time.Time, lookbackDays int, location *time.Location) (bool, error) {
	if lookbackDays <= 0 {
		return false, nil
	}
	for offset := 1; offset <= lookbackDays; offset++ {
		previousDay := day.AddDate(0, 0, -offset)
		entry, err := service.FetchLogByDate(userID, previousDay, location)
		if err != nil {
			return false, err
		}
		if entry.IsPeriod {
			return true, nil
		}
	}
	return false, nil
}
