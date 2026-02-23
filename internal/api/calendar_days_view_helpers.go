package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
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
	fertilityMap := make(map[string]bool)
	ovulationMap := make(map[string]bool)

	if !stats.FertilityWindowStart.IsZero() && !stats.FertilityWindowEnd.IsZero() {
		for day := stats.FertilityWindowStart; !day.After(stats.FertilityWindowEnd); day = day.AddDate(0, 0, 1) {
			fertilityMap[day.Format("2006-01-02")] = true
		}
	}
	if !stats.OvulationDate.IsZero() {
		ovulationMap[stats.OvulationDate.Format("2006-01-02")] = true
	}

	predictedCycleLength := int(stats.MedianCycleLength)
	if predictedCycleLength <= 0 {
		predictedCycleLength = int(stats.AverageCycleLength + 0.5)
	}
	if predictedCycleLength <= 0 {
		predictedCycleLength = models.DefaultCycleLength
	}

	predictedPeriodLength := int(stats.AveragePeriodLength + 0.5)
	if predictedPeriodLength <= 0 {
		predictedPeriodLength = models.DefaultPeriodLength
	}

	if !stats.NextPeriodStart.IsZero() {
		cycleStart := dateAtLocation(stats.NextPeriodStart, handler.location)
		for !cycleStart.After(gridEnd) {
			for offset := 0; offset < predictedPeriodLength; offset++ {
				day := cycleStart.AddDate(0, 0, offset)
				predictedPeriodMap[day.Format("2006-01-02")] = true
			}

			ovulationDate, fertilityStart, fertilityEnd, _, calculable := services.PredictCycleWindow(
				cycleStart,
				predictedCycleLength,
				predictedPeriodLength,
			)
			if calculable {
				ovulationMap[ovulationDate.Format("2006-01-02")] = true
				if !fertilityStart.IsZero() && !fertilityEnd.IsZero() {
					for day := fertilityStart; !day.After(fertilityEnd); day = day.AddDate(0, 0, 1) {
						fertilityMap[day.Format("2006-01-02")] = true
					}
				}
			}

			cycleStart = cycleStart.AddDate(0, 0, predictedCycleLength)
		}
	}

	todayKey := dateAtLocation(now, handler.location).Format("2006-01-02")

	days := make([]CalendarDay, 0, 42)
	for day := gridStart; !day.After(gridEnd); day = day.AddDate(0, 0, 1) {
		key := day.Format("2006-01-02")
		inMonth := day.Month() == monthStart.Month()
		entry, hasEntry := latestLogByDate[key]
		isPeriod := hasEntry && entry.IsPeriod
		isPredicted := predictedPeriodMap[key]
		isFertility := fertilityMap[key]
		isToday := key == todayKey
		isOvulation := ovulationMap[key]
		if isOvulation {
			isFertility = false
		}
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
		} else if isOvulation {
			cellClass += " calendar-cell-fertile"
			badgeClass += " calendar-tag-ovulation"
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
