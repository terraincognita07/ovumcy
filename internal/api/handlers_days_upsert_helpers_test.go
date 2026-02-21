package api

import (
	"errors"
	"strings"
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestNormalizeDayPayloadRejectsInvalidFlow(t *testing.T) {
	t.Parallel()

	_, err := normalizeDayPayload(dayPayload{
		IsPeriod: true,
		Flow:     "bad-flow",
	})
	if !errors.Is(err, errInvalidFlowValue) {
		t.Fatalf("expected errInvalidFlowValue, got %v", err)
	}
}

func TestNormalizeDayPayloadRequiresFlowForPeriodDay(t *testing.T) {
	t.Parallel()

	_, err := normalizeDayPayload(dayPayload{
		IsPeriod: true,
		Flow:     models.FlowNone,
	})
	if !errors.Is(err, errPeriodFlowRequired) {
		t.Fatalf("expected errPeriodFlowRequired, got %v", err)
	}
}

func TestNormalizeDayPayloadNormalizesNonPeriodFlowAndNotes(t *testing.T) {
	t.Parallel()

	payload := dayPayload{
		IsPeriod: false,
		Flow:     models.FlowHeavy,
		Notes:    strings.Repeat("x", maxDayNotesLength+15),
	}

	normalized, err := normalizeDayPayload(payload)
	if err != nil {
		t.Fatalf("normalize payload: %v", err)
	}
	if normalized.Flow != models.FlowNone {
		t.Fatalf("expected non-period flow normalized to %q, got %q", models.FlowNone, normalized.Flow)
	}
	if len(normalized.Notes) != maxDayNotesLength {
		t.Fatalf("expected notes length %d, got %d", maxDayNotesLength, len(normalized.Notes))
	}
}
