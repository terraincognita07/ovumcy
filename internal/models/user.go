package models

import "time"

const (
	RoleOwner   = "owner"
	RolePartner = "partner"
)

type User struct {
	ID           uint      `gorm:"primaryKey"`
	Email        string    `gorm:"uniqueIndex;not null"`
	PasswordHash string    `gorm:"not null"`
	Role         string    `gorm:"not null;default:owner"`
	CreatedAt    time.Time `gorm:"not null"`
}
