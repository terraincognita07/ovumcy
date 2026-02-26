package services

import (
	"errors"
	"strings"

	"github.com/terraincognita07/ovumcy/internal/models"
	"golang.org/x/crypto/bcrypt"
)

var ErrRecoveryCodeNotFound = errors.New("recovery code not found")

type AuthUserRepository interface {
	ExistsByNormalizedEmail(email string) (bool, error)
	FindByNormalizedEmail(email string) (models.User, error)
	FindByID(userID uint) (models.User, error)
	Create(user *models.User) error
	Save(user *models.User) error
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
