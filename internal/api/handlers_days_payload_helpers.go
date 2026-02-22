package api

import (
	"errors"

	"github.com/terraincognita07/lume/internal/models"
)

const maxDayNotesLength = 2000

var (
	errInvalidFlowValue = errors.New("invalid flow value")
)

func normalizeDayPayload(payload dayPayload) (dayPayload, error) {
	if !isValidFlow(payload.Flow) {
		return payload, errInvalidFlowValue
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
