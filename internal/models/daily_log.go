package models

import "time"

const (
	FlowNone   = "none"
	FlowLight  = "light"
	FlowMedium = "medium"
	FlowHeavy  = "heavy"
)

type DailyLog struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"not null;uniqueIndex:uidx_user_date"`
	Date       time.Time `gorm:"type:date;not null;uniqueIndex:uidx_user_date"`
	IsPeriod   bool      `gorm:"not null;default:false"`
	Flow       string    `gorm:"not null;default:none"`
	SymptomIDs []uint    `gorm:"serializer:json"`
	Notes      string
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
