package api

import (
	"time"

	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/services"
)

func (handler *Handler) buildCalendarDays(monthStart time.Time, logs []models.DailyLog, stats services.CycleStats, now time.Time) []CalendarDay {
	monthEnd := monthStart.AddDate(0, 1, -1)
	gridStart := monthStart.AddDate(0, 0, -int(monthStart.Weekday()))
	gridEnd := monthEnd.AddDate(0, 0, 6-int(monthEnd.Weekday()))

	latestLogByDate := make(map[string]models.DailyLog)
	hasDataMap := make(map[string]bool)
	for _, logEntry := range logs {
		key := dateAtLocation(logEntry.Date, handler.location).Format("2006-01-02")
		existing, exists := latestLogByDate[key]
		if !exists || logEntry.Date.After(existing.Date) || (logEntry.Date.Equal(existing.Date) && logEntry.ID > existing.ID) {
			latestLogByDate[key] = logEntry
		}
		hasDataMap[key] = hasDataMap[key] || dayHasData(logEntry)
	}

	predictedPeriodMap := make(map[string]bool)
	predictedPeriodLength := int(stats.AveragePeriodLength + 0.5)
	if predictedPeriodLength <= 0 {
		predictedPeriodLength = 5
	}
	if !stats.NextPeriodStart.IsZero() {
		for offset := 0; offset < predictedPeriodLength; offset++ {
			day := stats.NextPeriodStart.AddDate(0, 0, offset)
			predictedPeriodMap[day.Format("2006-01-02")] = true
		}
	}

	fertilityMap := make(map[string]bool)
	if !stats.FertilityWindowStart.IsZero() && !stats.FertilityWindowEnd.IsZero() {
		for day := stats.FertilityWindowStart; !day.After(stats.FertilityWindowEnd); day = day.AddDate(0, 0, 1) {
			fertilityMap[day.Format("2006-01-02")] = true
		}
	}

	todayKey := dateAtLocation(now, handler.location).Format("2006-01-02")
	ovulationKey := stats.OvulationDate.Format("2006-01-02")

	days := make([]CalendarDay, 0, 42)
	for day := gridStart; !day.After(gridEnd); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		inMonth := day.Month() == monthStart.Month()
		entry, hasEntry := latestLogByDate[key]
		isPeriod := hasEntry && entry.IsPeriod
		isPredicted := predictedPeriodMap[key]
		isFertility := fertilityMap[key]
		isToday := key == todayKey
		isOvulation := key == ovulationKey
		hasData := hasDataMap[key]

		cellClass := "calendar-cell"
		textClass := "calendar-day-number"
		badgeClass := "calendar-tag"
		if isPeriod {
			cellClass += " calendar-cell-period"
			badgeClass += " calendar-tag-period"
		} else if isPredicted {
			cellClass += " calendar-cell-predicted"
			badgeClass += " calendar-tag-predicted"
		} else if isFertility {
			cellClass += " calendar-cell-fertile"
			badgeClass += " calendar-tag-fertile"
		}
		if !inMonth {
			cellClass += " calendar-cell-out"
			textClass += " calendar-day-out"
		}
		if isToday {
			cellClass += " calendar-cell-today"
		}

		days = append(days, CalendarDay{
			Date:         day,
			DateString:   key,
			Day:          day.Day(),
			InMonth:      inMonth,
			IsToday:      isToday,
			IsPeriod:     isPeriod,
			IsPredicted:  isPredicted,
			IsFertility:  isFertility,
			IsOvulation:  isOvulation,
			HasData:      hasData,
			CellClass:    cellClass,
			TextClass:    textClass,
			BadgeClass:   badgeClass,
			OvulationDot: isOvulation,
		})
	}
	return days
}
