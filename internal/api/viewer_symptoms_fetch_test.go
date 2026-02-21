package api

import (
	"testing"

	"github.com/terraincognita07/lume/internal/models"
)

func TestFetchSymptomsForViewer_NonOwnerReturnsEmpty(t *testing.T) {
	handler := &Handler{}
	partner := &models.User{Role: models.RolePartner}

	symptoms, err := handler.fetchSymptomsForViewer(partner)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(symptoms) != 0 {
		t.Fatalf("expected empty symptoms for non-owner, got %#v", symptoms)
	}
}

func TestFetchSymptomsForViewer_NilUserReturnsEmpty(t *testing.T) {
	handler := &Handler{}

	symptoms, err := handler.fetchSymptomsForViewer(nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(symptoms) != 0 {
		t.Fatalf("expected empty symptoms for nil user, got %#v", symptoms)
	}
}
