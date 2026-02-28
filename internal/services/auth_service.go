package services

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/terraincognita07/ovumcy/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrRecoveryCodeNotFound = errors.New("recovery code not found")
	ErrInvalidResetToken    = errors.New("invalid reset token")
	ErrAuthUserRequired     = errors.New("auth user is required")
	ErrRecoveryCodeGenerate = errors.New("recovery code generation failed")
	ErrRecoveryCodeUpdate   = errors.New("recovery code update failed")
)

type AuthUserRepository interface {
	ExistsByNormalizedEmail(email string) (bool, error)
	FindByNormalizedEmail(email string) (models.User, error)
	FindByID(userID uint) (models.User, error)
	Create(user *models.User) error
	Save(user *models.User) error
	UpdateRecoveryCodeHash(userID uint, recoveryHash string) error
	ListWithRecoveryCodeHash() ([]models.User, error)
}

type AuthService struct {
	users AuthUserRepository
}

func NewAuthService(users AuthUserRepository) *AuthService {
	return &AuthService{users: users}
}

func (service *AuthService) RegistrationEmailExists(email string) (bool, error) {
	return service.users.ExistsByNormalizedEmail(email)
}

func (service *AuthService) CreateUser(user *models.User) error {
	return service.users.Create(user)
}

func (service *AuthService) FindByNormalizedEmail(email string) (models.User, error) {
	return service.users.FindByNormalizedEmail(email)
}

func (service *AuthService) FindByID(userID uint) (models.User, error) {
	return service.users.FindByID(userID)
}

func (service *AuthService) SaveUser(user *models.User) error {
	return service.users.Save(user)
}

func (service *AuthService) FindUserByRecoveryCode(code string) (*models.User, error) {
	users, err := service.users.ListWithRecoveryCodeHash()
	if err != nil {
		return nil, err
	}

	for index := range users {
		hash := strings.TrimSpace(users[index].RecoveryCodeHash)
		if hash == "" {
			continue
		}
		if bcrypt.CompareHashAndPassword([]byte(hash), []byte(code)) == nil {
			return &users[index], nil
		}
	}
	return nil, ErrRecoveryCodeNotFound
}

func (service *AuthService) BuildPasswordResetToken(secretKey []byte, userID uint, passwordHash string, ttl time.Duration, now time.Time) (string, error) {
	return BuildPasswordResetToken(secretKey, userID, passwordHash, ttl, now)
}

func (service *AuthService) ResolveUserByResetToken(secretKey []byte, rawToken string, now time.Time) (*models.User, error) {
	claims, err := ParsePasswordResetToken(secretKey, rawToken, now)
	if err != nil {
		return nil, ErrInvalidResetToken
	}

	user, err := service.users.FindByID(claims.UserID)
	if err != nil {
		return nil, ErrInvalidResetToken
	}
	if !IsPasswordStateFingerprintMatch(claims.PasswordState, user.PasswordHash) {
		return nil, ErrInvalidResetToken
	}
	return &user, nil
}

func (service *AuthService) GenerateRecoveryCodeHash() (string, string, error) {
	return GenerateRecoveryCodeHash()
}

func (service *AuthService) RegenerateRecoveryCode(userID uint) (string, error) {
	recoveryCode, recoveryHash, err := GenerateRecoveryCodeHash()
	if err != nil {
		return "", fmt.Errorf("%w: %v", ErrRecoveryCodeGenerate, err)
	}
	if err := service.users.UpdateRecoveryCodeHash(userID, recoveryHash); err != nil {
		return "", fmt.Errorf("%w: %v", ErrRecoveryCodeUpdate, err)
	}
	return recoveryCode, nil
}

func (service *AuthService) ResetPasswordAndRotateRecoveryCode(user *models.User, newPassword string) (string, error) {
	if user == nil {
		return "", ErrAuthUserRequired
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	recoveryCode, recoveryHash, err := GenerateRecoveryCodeHash()
	if err != nil {
		return "", err
	}

	user.PasswordHash = string(passwordHash)
	user.RecoveryCodeHash = recoveryHash
	user.MustChangePassword = false
	if err := service.users.Save(user); err != nil {
		return "", err
	}

	return recoveryCode, nil
}
