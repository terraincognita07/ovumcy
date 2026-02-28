package services

import (
	"errors"
	"testing"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"golang.org/x/crypto/bcrypt"
)

type stubAuthUserRepo struct {
	user                  models.User
	findByIDErr           error
	saveErr               error
	updateRecoveryCodeErr error
	updatedUserID         uint
	updatedRecoveryHash   string
	saveCalled            bool
}

func (stub *stubAuthUserRepo) ExistsByNormalizedEmail(string) (bool, error) {
	return false, nil
}

func (stub *stubAuthUserRepo) FindByNormalizedEmail(string) (models.User, error) {
	return models.User{}, errors.New("not implemented")
}

func (stub *stubAuthUserRepo) FindByID(uint) (models.User, error) {
	if stub.findByIDErr != nil {
		return models.User{}, stub.findByIDErr
	}
	return stub.user, nil
}

func (stub *stubAuthUserRepo) Create(*models.User) error {
	return errors.New("not implemented")
}

func (stub *stubAuthUserRepo) Save(user *models.User) error {
	if stub.saveErr != nil {
		return stub.saveErr
	}
	stub.saveCalled = true
	stub.user = *user
	return nil
}

func (stub *stubAuthUserRepo) UpdateRecoveryCodeHash(userID uint, recoveryHash string) error {
	if stub.updateRecoveryCodeErr != nil {
		return stub.updateRecoveryCodeErr
	}
	stub.updatedUserID = userID
	stub.updatedRecoveryHash = recoveryHash
	return nil
}

func (stub *stubAuthUserRepo) ListWithRecoveryCodeHash() ([]models.User, error) {
	return []models.User{stub.user}, nil
}

func TestAuthServiceResolveUserByResetToken(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}

	repo := &stubAuthUserRepo{
		user: models.User{
			ID:           42,
			PasswordHash: string(passwordHash),
		},
	}
	service := NewAuthService(repo)

	token, err := service.BuildPasswordResetToken(secret, 42, repo.user.PasswordHash, 30*time.Minute, now)
	if err != nil {
		t.Fatalf("BuildPasswordResetToken() unexpected error: %v", err)
	}

	user, err := service.ResolveUserByResetToken(secret, token, now.Add(1*time.Minute))
	if err != nil {
		t.Fatalf("ResolveUserByResetToken() unexpected error: %v", err)
	}
	if user.ID != 42 {
		t.Fatalf("expected user id 42, got %d", user.ID)
	}
}

func TestAuthServiceResolveUserByResetTokenRejectsStateMismatch(t *testing.T) {
	secret := []byte("test-secret")
	now := time.Date(2026, time.March, 1, 10, 0, 0, 0, time.UTC)

	originalHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash original password: %v", err)
	}
	changedHash, err := bcrypt.GenerateFromPassword([]byte("DifferentPass2"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash changed password: %v", err)
	}

	repo := &stubAuthUserRepo{
		user: models.User{
			ID:           42,
			PasswordHash: string(changedHash),
		},
	}
	service := NewAuthService(repo)
	token, err := service.BuildPasswordResetToken(secret, 42, string(originalHash), 30*time.Minute, now)
	if err != nil {
		t.Fatalf("BuildPasswordResetToken() unexpected error: %v", err)
	}

	if _, err := service.ResolveUserByResetToken(secret, token, now.Add(1*time.Minute)); !errors.Is(err, ErrInvalidResetToken) {
		t.Fatalf("expected ErrInvalidResetToken, got %v", err)
	}
}

func TestAuthServiceResetPasswordAndRotateRecoveryCode(t *testing.T) {
	originalHash, err := bcrypt.GenerateFromPassword([]byte("StrongPass1"), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("hash original password: %v", err)
	}

	user := models.User{
		ID:                 7,
		PasswordHash:       string(originalHash),
		RecoveryCodeHash:   "old-hash",
		MustChangePassword: true,
	}
	repo := &stubAuthUserRepo{user: user}
	service := NewAuthService(repo)

	recoveryCode, err := service.ResetPasswordAndRotateRecoveryCode(&user, "EvenStronger2")
	if err != nil {
		t.Fatalf("ResetPasswordAndRotateRecoveryCode() unexpected error: %v", err)
	}
	if recoveryCode == "" {
		t.Fatalf("expected non-empty recovery code")
	}
	if !repo.saveCalled {
		t.Fatalf("expected Save() to be called")
	}
	if repo.user.MustChangePassword {
		t.Fatalf("expected MustChangePassword=false after reset")
	}
	if repo.user.RecoveryCodeHash == "" || repo.user.RecoveryCodeHash == "old-hash" {
		t.Fatalf("expected rotated recovery code hash")
	}
	if bcrypt.CompareHashAndPassword([]byte(repo.user.PasswordHash), []byte("EvenStronger2")) != nil {
		t.Fatalf("expected password hash updated to new password")
	}
}

func TestAuthServiceRegenerateRecoveryCode(t *testing.T) {
	repo := &stubAuthUserRepo{}
	service := NewAuthService(repo)

	recoveryCode, err := service.RegenerateRecoveryCode(55)
	if err != nil {
		t.Fatalf("RegenerateRecoveryCode() unexpected error: %v", err)
	}
	if recoveryCode == "" {
		t.Fatalf("expected non-empty recovery code")
	}
	if repo.updatedUserID != 55 {
		t.Fatalf("expected UpdateRecoveryCodeHash to be called for user 55, got %d", repo.updatedUserID)
	}
	if repo.updatedRecoveryHash == "" {
		t.Fatalf("expected non-empty recovery hash update")
	}
}
