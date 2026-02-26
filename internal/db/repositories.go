package db

import "gorm.io/gorm"

type Repositories struct {
	Users     *UserRepository
	DailyLogs *DailyLogRepository
	Symptoms  *SymptomRepository
}

func NewRepositories(database *gorm.DB) *Repositories {
	return &Repositories{
		Users:     NewUserRepository(database),
		DailyLogs: NewDailyLogRepository(database),
		Symptoms:  NewSymptomRepository(database),
	}
}
