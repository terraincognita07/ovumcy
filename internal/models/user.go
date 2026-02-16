package models

import "time"

const (
	RoleOwner   = "owner"
	RolePartner = "partner"
)

type User struct {
	ID                 uint      `gorm:"primaryKey"`
	Email              string    `gorm:"uniqueIndex;not null"`
	PasswordHash       string    `gorm:"not null"`
	RecoveryCodeHash   string    `gorm:"column:recovery_code_hash"`
	MustChangePassword bool      `gorm:"column:must_change_password;not null;default:false"`
	Role               string    `gorm:"not null;default:owner"`
	CreatedAt          time.Time `gorm:"not null"`
}
