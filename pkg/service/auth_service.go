package service

import (
	"fmt"
	"social-network/pkg/models"
	"social-network/pkg/repository"
	"social-network/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type AuthService interface {
	UserRegister(request *models.RegisterRequest) (*models.Profile, error)
	Login(userId, password string) (*models.User, error)
}

type authService struct {
	userRepository    repository.UserRepository
	profileRepository repository.ProfileRepository
}

func InitAuthService(userRepository repository.UserRepository, profileRepository repository.ProfileRepository) AuthService {
	return &authService{
		userRepository:    userRepository,
		profileRepository: profileRepository,
	}
}

func (authService *authService) UserRegister(request *models.RegisterRequest) (*models.Profile, error) {
	err := utils.ValidateRegisterRequest(request.FirstName, request.LastName, request.Password, request.Gender, request.Biography, request.City)
	if err != nil {
		return nil, err
	}

	birthDate, err := time.Parse("2006-01-02", request.Birthdate)
	if err != nil {
		return nil, fmt.Errorf("Дата рождения указана некорректно")
	}
	user := models.User{
		Id: uuid.New(),
	}

	pass, err := utils.HashPassword(request.Password)
	if err != nil {
		return nil, err
	}

	err = authService.userRepository.Create(&user, pass)
	if err != nil {
		return nil, err
	}

	profile := models.Profile{
		Id:        uuid.New(),
		UserId:    user.Id,
		FirstName: request.FirstName,
		LastName:  request.LastName,
		Birthdate: birthDate,
		Gender:    request.Gender,
		Biography: request.Biography,
		City:      request.City,
	}

	err = authService.profileRepository.Create(&profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (authService *authService) Login(userId, password string) (*models.User, error) {
	id, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("Пользователь не зарегистрирован")
	}

	user, err := authService.userRepository.GetUserById(id)

	if err != nil {
		return nil, fmt.Errorf("Пользователь не зарегистрирован")
	}

	isValidPassword := utils.CheckPassword(password, user.Password)

	if !isValidPassword {
		return nil, fmt.Errorf("Логин или пароль указан неверно")
	}

	return user, nil
}
