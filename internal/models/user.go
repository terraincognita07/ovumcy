package models

import "time"

const (
	RoleOwner   = "owner"
	RolePartner = "partner"
)

type User struct {
	ID                  uint       `gorm:"primaryKey"`
	Email               string     `gorm:"uniqueIndex;not null"`
	PasswordHash        string     `gorm:"not null"`
	RecoveryCodeHash    string     `gorm:"column:recovery_code_hash"`
	MustChangePassword  bool       `gorm:"column:must_change_password;not null;default:false"`
	Role                string     `gorm:"not null;default:owner"`
	OnboardingCompleted bool       `gorm:"not null;default:false"`
	CycleLength         int        `gorm:"not null;default:28"`
	PeriodLength        int        `gorm:"not null;default:5"`
	LastPeriodStart     *time.Time `gorm:"type:date"`
	CreatedAt           time.Time  `gorm:"not null"`
}
