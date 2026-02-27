package services

import (
	"errors"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
)

type dayLogRepositoryStub struct {
	entries        map[string]models.DailyLog
	nextID         uint
	findErrByDay   map[string]error
	createErrByDay map[string]error
	saveErrByDay   map[string]error
}

func newDayLogRepositoryStub() *dayLogRepositoryStub {
	return &dayLogRepositoryStub{
		entries:        make(map[string]models.DailyLog),
		nextID:         1,
		findErrByDay:   make(map[string]error),
		createErrByDay: make(map[string]error),
		saveErrByDay:   make(map[string]error),
	}
}

func (stub *dayLogRepositoryStub) dayKey(value time.Time) string {
	return value.Format("2006-01-02")
}

func (stub *dayLogRepositoryStub) ListByUser(userID uint) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	for _, entry := range stub.entries {
		if entry.UserID == userID {
			logs = append(logs, entry)
		}
	}
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].Date.Equal(logs[j].Date) {
			return logs[i].ID < logs[j].ID
		}
		return logs[i].Date.Before(logs[j].Date)
	})
	return logs, nil
}

func (stub *dayLogRepositoryStub) ListByUserRange(userID uint, fromStart *time.Time, toEnd *time.Time) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	for _, entry := range stub.entries {
		if entry.UserID != userID {
			continue
		}
		if fromStart != nil && entry.Date.Before(*fromStart) {
			continue
		}
		if toEnd != nil && !entry.Date.Before(*toEnd) {
			continue
		}
		logs = append(logs, entry)
	}
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].Date.Equal(logs[j].Date) {
			return logs[i].ID < logs[j].ID
		}
		return logs[i].Date.Before(logs[j].Date)
	})
	return logs, nil
}

func (stub *dayLogRepositoryStub) ListByUserDayRange(userID uint, dayStart time.Time, dayEnd time.Time) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	for _, entry := range stub.entries {
		if entry.UserID != userID {
			continue
		}
		if entry.Date.Before(dayStart) || !entry.Date.Before(dayEnd) {
			continue
		}
		logs = append(logs, entry)
	}
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].Date.Equal(logs[j].Date) {
			return logs[i].ID > logs[j].ID
		}
		return logs[i].Date.After(logs[j].Date)
	})
	return logs, nil
}

func (stub *dayLogRepositoryStub) ListPeriodDays(userID uint) ([]models.DailyLog, error) {
	logs := make([]models.DailyLog, 0)
	for _, entry := range stub.entries {
		if entry.UserID == userID && entry.IsPeriod {
			logs = append(logs, models.DailyLog{
				Date:     entry.Date,
				IsPeriod: true,
			})
		}
	}
	sort.Slice(logs, func(i, j int) bool {
		return logs[i].Date.Before(logs[j].Date)
	})
	return logs, nil
}

func (stub *dayLogRepositoryStub) FindByUserAndDayRange(userID uint, dayStart time.Time, dayEnd time.Time) (models.DailyLog, bool, error) {
	key := stub.dayKey(dayStart)
	if err, ok := stub.findErrByDay[key]; ok {
		return models.DailyLog{}, false, err
	}

	entry, ok := stub.entries[key]
	if !ok || entry.UserID != userID || entry.Date.Before(dayStart) || !entry.Date.Before(dayEnd) {
		return models.DailyLog{}, false, nil
	}
	return entry, true, nil
}

func (stub *dayLogRepositoryStub) Create(entry *models.DailyLog) error {
	key := stub.dayKey(entry.Date)
	if err, ok := stub.createErrByDay[key]; ok {
		return err
	}
	if entry.ID == 0 {
		entry.ID = stub.nextID
		stub.nextID++
	}
	stub.entries[key] = *entry
	return nil
}

func (stub *dayLogRepositoryStub) Save(entry *models.DailyLog) error {
	key := stub.dayKey(entry.Date)
	if err, ok := stub.saveErrByDay[key]; ok {
		return err
	}
	stub.entries[key] = *entry
	return nil
}

func (stub *dayLogRepositoryStub) DeleteByUserAndDayRange(userID uint, dayStart time.Time, dayEnd time.Time) error {
	for key, entry := range stub.entries {
		if entry.UserID != userID {
			continue
		}
		if entry.Date.Before(dayStart) || !entry.Date.Before(dayEnd) {
			continue
		}
		delete(stub.entries, key)
	}
	return nil
}

type dayUserRepositoryStub struct {
	settings  models.User
	loadErr   error
	updateErr error
	updates   []map[string]any
}

func (stub *dayUserRepositoryStub) LoadSettingsByID(uint) (models.User, error) {
	if stub.loadErr != nil {
		return models.User{}, stub.loadErr
	}
	return stub.settings, nil
}

func (stub *dayUserRepositoryStub) UpdateByID(_ uint, updates map[string]any) error {
	if stub.updateErr != nil {
		return stub.updateErr
	}
	copied := make(map[string]any, len(updates))
	for key, value := range updates {
		copied[key] = value
	}
	stub.updates = append(stub.updates, copied)
	return nil
}

func TestUpsertDayEntryWithAutoFillNormalizesNonPeriodInput(t *testing.T) {
	logs := newDayLogRepositoryStub()
	users := &dayUserRepositoryStub{}
	service := NewDayService(logs, users)

	entry, err := service.UpsertDayEntryWithAutoFill(
		10,
		time.Date(2026, time.February, 20, 12, 0, 0, 0, time.UTC),
		DayEntryInput{
			IsPeriod:   false,
			Flow:       models.FlowHeavy,
			SymptomIDs: []uint{5, 6},
			Notes:      strings.Repeat("x", MaxDayNotesLength+11),
		},
		time.UTC,
	)
	if err != nil {
		t.Fatalf("UpsertDayEntryWithAutoFill() unexpected error: %v", err)
	}
	if entry.Flow != models.FlowNone {
		t.Fatalf("expected non-period flow normalized to %q, got %q", models.FlowNone, entry.Flow)
	}
	if len(entry.SymptomIDs) != 0 {
		t.Fatalf("expected non-period symptom IDs to be cleared, got %#v", entry.SymptomIDs)
	}
	if len(entry.Notes) != MaxDayNotesLength {
		t.Fatalf("expected notes length %d, got %d", MaxDayNotesLength, len(entry.Notes))
	}
	if len(users.updates) != 1 {
		t.Fatalf("expected one last_period_start sync call, got %d", len(users.updates))
	}
	if _, ok := users.updates[0]["last_period_start"]; !ok {
		t.Fatalf("expected last_period_start key in sync update, got %#v", users.updates[0])
	}
	if users.updates[0]["last_period_start"] != nil {
		t.Fatalf("expected last_period_start nil for non-period logs, got %#v", users.updates[0]["last_period_start"])
	}
}

func TestUpsertDayEntryWithAutoFillCreatesFollowingPeriodDays(t *testing.T) {
	logs := newDayLogRepositoryStub()
	users := &dayUserRepositoryStub{
		settings: models.User{
			PeriodLength:   3,
			AutoPeriodFill: true,
		},
	}
	service := NewDayService(logs, users)

	day := time.Date(2026, time.February, 10, 8, 0, 0, 0, time.UTC)
	entry, err := service.UpsertDayEntryWithAutoFill(
		10,
		day,
		DayEntryInput{
			IsPeriod: true,
			Flow:     models.FlowLight,
			Notes:    "period",
		},
		time.UTC,
	)
	if err != nil {
		t.Fatalf("UpsertDayEntryWithAutoFill() unexpected error: %v", err)
	}
	if !entry.IsPeriod {
		t.Fatalf("expected created entry to be period day")
	}

	expectedDays := []string{"2026-02-10", "2026-02-11", "2026-02-12"}
	for _, dayKey := range expectedDays {
		logEntry, ok := logs.entries[dayKey]
		if !ok {
			t.Fatalf("expected day %s to exist after autofill", dayKey)
		}
		if !logEntry.IsPeriod {
			t.Fatalf("expected day %s to be period", dayKey)
		}
	}

	if len(users.updates) != 1 {
		t.Fatalf("expected one sync update, got %d", len(users.updates))
	}
	gotStart, ok := users.updates[0]["last_period_start"].(time.Time)
	if !ok {
		t.Fatalf("expected time.Time last_period_start, got %#v", users.updates[0]["last_period_start"])
	}
	if gotStart.Format("2006-01-02") != "2026-02-10" {
		t.Fatalf("expected last_period_start 2026-02-10, got %s", gotStart.Format("2006-01-02"))
	}
}

func TestUpsertDayEntryWithAutoFillReturnsTypedLoadError(t *testing.T) {
	logs := newDayLogRepositoryStub()
	users := &dayUserRepositoryStub{loadErr: errors.New("load settings failed")}
	service := NewDayService(logs, users)

	_, err := service.UpsertDayEntryWithAutoFill(
		10,
		time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC),
		DayEntryInput{
			IsPeriod: true,
			Flow:     models.FlowLight,
		},
		time.UTC,
	)
	if !errors.Is(err, ErrDayAutoFillLoadFailed) {
		t.Fatalf("expected ErrDayAutoFillLoadFailed, got %v", err)
	}
}

func TestUpsertDayEntryWithAutoFillReturnsTypedAutofillDecisionError(t *testing.T) {
	logs := newDayLogRepositoryStub()
	logs.findErrByDay["2026-02-09"] = errors.New("previous day read failed")
	users := &dayUserRepositoryStub{
		settings: models.User{
			PeriodLength:   3,
			AutoPeriodFill: true,
		},
	}
	service := NewDayService(logs, users)

	_, err := service.UpsertDayEntryWithAutoFill(
		10,
		time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC),
		DayEntryInput{
			IsPeriod: true,
			Flow:     models.FlowLight,
		},
		time.UTC,
	)
	if !errors.Is(err, ErrDayAutoFillCheckFailed) {
		t.Fatalf("expected ErrDayAutoFillCheckFailed, got %v", err)
	}
}

func TestUpsertDayEntryWithAutoFillReturnsTypedAutofillApplyError(t *testing.T) {
	logs := newDayLogRepositoryStub()
	logs.createErrByDay["2026-02-11"] = errors.New("autofill create failed")
	users := &dayUserRepositoryStub{
		settings: models.User{
			PeriodLength:   3,
			AutoPeriodFill: true,
		},
	}
	service := NewDayService(logs, users)

	_, err := service.UpsertDayEntryWithAutoFill(
		10,
		time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC),
		DayEntryInput{
			IsPeriod: true,
			Flow:     models.FlowLight,
		},
		time.UTC,
	)
	if !errors.Is(err, ErrDayAutoFillApplyFailed) {
		t.Fatalf("expected ErrDayAutoFillApplyFailed, got %v", err)
	}
}

func TestUpsertDayEntryWithAutoFillReturnsTypedSyncError(t *testing.T) {
	logs := newDayLogRepositoryStub()
	users := &dayUserRepositoryStub{updateErr: errors.New("sync failed")}
	service := NewDayService(logs, users)

	_, err := service.UpsertDayEntryWithAutoFill(
		10,
		time.Date(2026, time.February, 10, 0, 0, 0, 0, time.UTC),
		DayEntryInput{
			IsPeriod: false,
			Flow:     models.FlowNone,
		},
		time.UTC,
	)
	if !errors.Is(err, ErrSyncLastPeriodFailed) {
		t.Fatalf("expected ErrSyncLastPeriodFailed, got %v", err)
	}
}
