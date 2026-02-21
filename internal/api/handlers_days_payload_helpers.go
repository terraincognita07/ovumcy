package api

import "github.com/terraincognita07/lume/internal/models"

const maxDayNotesLength = 2000

func normalizeDayPayload(payload dayPayload) (dayPayload, error) {
	if !isValidFlow(payload.Flow) {
		return payload, errInvalidFlowValue
	}
	if payload.IsPeriod && payload.Flow == models.FlowNone {
		return payload, errPeriodFlowRequired
	}
	if !payload.IsPeriod {
		payload.Flow = models.FlowNone
	}
	payload.Notes = trimDayNotes(payload.Notes)
	return payload, nil
}

func trimDayNotes(value string) string {
	if len(value) > maxDayNotesLength {
		return value[:maxDayNotesLength]
	}
	return value
}
