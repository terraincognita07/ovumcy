package api

import (
	"errors"
	"testing"
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"gorm.io/gorm"
)

func assertOnboardingPeriodLogForDay(t *testing.T, database *gorm.DB, userID uint, day time.Time) {
	t.Helper()

	entry, err := findOnboardingLogByDay(database, userID, day)
	if err != nil {
		t.Fatalf("expected onboarding log for %s: %v", day.Format("2006-01-02"), err)
	}
	if !entry.IsPeriod {
		t.Fatalf("expected %s to be marked as period day", day.Format("2006-01-02"))
	}
	if entry.Flow != models.FlowNone {
		t.Fatalf("expected flow=none for %s, got %q", day.Format("2006-01-02"), entry.Flow)
	}
}

func assertNoOnboardingLogForDay(t *testing.T, database *gorm.DB, userID uint, day time.Time) {
	t.Helper()

	_, err := findOnboardingLogByDay(database, userID, day)
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		t.Fatalf("expected no onboarding log for %s, got err=%v", day.Format("2006-01-02"), err)
	}
}

func findOnboardingLogByDay(database *gorm.DB, userID uint, day time.Time) (models.DailyLog, error) {
	dayStart := time.Date(day.Year(), day.Month(), day.Day(), 0, 0, 0, 0, day.Location())
	dayEnd := dayStart.AddDate(0, 0, 1)
	dayKeys := []string{
		dayStart.Format("2006-01-02"),
		dayStart.UTC().Format("2006-01-02"),
		dayEnd.Add(-time.Nanosecond).UTC().Format("2006-01-02"),
	}

	candidates := make([]models.DailyLog, 0)
	err := database.
		Where(
			"user_id = ? AND substr(date, 1, 10) IN ?",
			userID,
			dayKeys,
		).
		Order("date DESC, id DESC").
		Find(&candidates).Error
	if err != nil {
		return models.DailyLog{}, err
	}

	for _, candidate := range candidates {
		candidateLocal := candidate.Date.In(day.Location())
		candidateStart := time.Date(candidateLocal.Year(), candidateLocal.Month(), candidateLocal.Day(), 0, 0, 0, 0, day.Location())
		if candidateStart.Equal(dayStart) {
			return candidate, nil
		}
	}

	return models.DailyLog{}, gorm.ErrRecordNotFound
}
