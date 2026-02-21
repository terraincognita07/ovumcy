package cli

import (
	"errors"
	"fmt"
	"net/mail"
	"strings"

	"github.com/terraincognita07/lume/internal/db"
	"github.com/terraincognita07/lume/internal/models"
	"github.com/terraincognita07/lume/internal/security"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func RunResetPasswordCommand(dbPath string, email string) error {
	normalizedEmail := strings.ToLower(strings.TrimSpace(email))
	if normalizedEmail == "" {
		return errors.New("email is required")
	}
	if _, err := mail.ParseAddress(normalizedEmail); err != nil {
		return fmt.Errorf("invalid email address: %w", err)
	}

	database, err := db.OpenSQLite(dbPath)
	if err != nil {
		return fmt.Errorf("database init failed: %w", err)
	}

	var user models.User
	if err := database.Where("email = ?", normalizedEmail).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user %s not found", normalizedEmail)
		}
		return fmt.Errorf("load user: %w", err)
	}

	temporaryPassword, err := generateTemporaryPassword(12)
	if err != nil {
		return fmt.Errorf("generate temporary password: %w", err)
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(temporaryPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("hash temporary password: %w", err)
	}

	user.PasswordHash = string(passwordHash)
	user.MustChangePassword = true
	if err := database.Save(&user).Error; err != nil {
		return fmt.Errorf("update user password: %w", err)
	}

	fmt.Println("✅ Password reset successful")
	fmt.Printf("Temporary password: %s\n", temporaryPassword)
	fmt.Println("User must change password on next login.")

	return nil
}

func generateTemporaryPassword(length int) (string, error) {
	if length < 8 {
		length = 8
	}

	const alphabet = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789"
	return security.RandomString(length, alphabet)
}
