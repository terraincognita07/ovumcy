package services

type SetupUserRepository interface {
	CountUsers() (int64, error)
}

type SetupService struct {
	users SetupUserRepository
}

func NewSetupService(users SetupUserRepository) *SetupService {
	return &SetupService{users: users}
}

func (service *SetupService) RequiresInitialSetup() (bool, error) {
	usersCount, err := service.users.CountUsers()
	if err != nil {
		return false, err
	}
	return usersCount == 0, nil
}
