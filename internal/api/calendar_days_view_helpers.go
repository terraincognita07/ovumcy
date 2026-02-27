package api

import (
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"github.com/terraincognita07/ovumcy/internal/services"
)

func (handler *Handler) buildCalendarDays(monthStart time.Time, logs []models.DailyLog, stats services.CycleStats, now time.Time) []CalendarDay {
	states := services.BuildCalendarDayStates(monthStart, logs, stats, now, handler.location)
	days := make([]CalendarDay, 0, len(states))
	for _, state := range states {
		cellClass := "calendar-cell"
		textClass := "calendar-day-number"
		badgeClass := "calendar-tag"
		if state.IsPeriod {
			cellClass += " calendar-cell-period"
			badgeClass += " calendar-tag-period"
		} else if state.IsPredicted {
			cellClass += " calendar-cell-predicted"
			badgeClass += " calendar-tag-predicted"
		} else if state.IsOvulation {
			cellClass += " calendar-cell-fertile"
			badgeClass += " calendar-tag-ovulation"
		} else if state.IsFertility {
			cellClass += " calendar-cell-fertile"
			badgeClass += " calendar-tag-fertile"
		}
		if !state.InMonth {
			cellClass += " calendar-cell-out"
			textClass += " calendar-day-out"
		}
		if state.IsToday {
			cellClass += " calendar-cell-today"
		}

		days = append(days, CalendarDay{
			Date:         state.Date,
			DateString:   state.DateString,
			Day:          state.Day,
			InMonth:      state.InMonth,
			IsToday:      state.IsToday,
			IsPeriod:     state.IsPeriod,
			IsPredicted:  state.IsPredicted,
			IsFertility:  state.IsFertility,
			IsOvulation:  state.IsOvulation,
			HasData:      state.HasData,
			CellClass:    cellClass,
			TextClass:    textClass,
			BadgeClass:   badgeClass,
			OvulationDot: state.IsOvulation,
		})
	}
	return days
}
