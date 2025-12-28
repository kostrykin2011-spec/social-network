package service

import (
	"context"
	"fmt"
	"social-network/internal/config"
	"social-network/pkg/database"
	"social-network/pkg/models"
	"social-network/pkg/repository"
	"social-network/pkg/utils"
	"time"

	"github.com/google/uuid"
)

type AuthService interface {
	UserRegister(ctx context.Context, request *models.RegisterRequest) (*models.Profile, error)
	Login(ctx context.Context, userId, password string) (*models.AuthResponse, error)
}

type authService struct {
	config            *config.Config
	userRepository    repository.UserRepository
	profileRepository repository.ProfileRepository
}

func InitAuthService(config *config.Config, userRepository repository.UserRepository, profileRepository repository.ProfileRepository) AuthService {
	return &authService{
		config:            config,
		userRepository:    userRepository,
		profileRepository: profileRepository,
	}
}

func (authService *authService) UserRegister(ctx context.Context, request *models.RegisterRequest) (*models.Profile, error) {
	ctx = database.WithMaster(ctx)
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

	err = authService.userRepository.Create(ctx, &user, pass)
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

	err = authService.profileRepository.Create(ctx, &profile)
	if err != nil {
		return nil, err
	}

	return &profile, nil
}

func (authService *authService) Login(ctx context.Context, userId, password string) (*models.AuthResponse, error) {
	ctx = database.WithReplica(ctx)
	id, err := uuid.Parse(userId)
	if err != nil {
		return nil, fmt.Errorf("Пользователь не зарегистрирован")
	}

	user, err := authService.userRepository.GetUserById(ctx, id)

	if err != nil {
		return nil, fmt.Errorf("Пользователь не зарегистрирован")
	}

	isValidPassword := utils.CheckPassword(password, user.Password)

	if !isValidPassword {
		return nil, fmt.Errorf("Логин или пароль указан неверно")
	}

	token, err := utils.GenerateToken(user.Id, authService.config)

	return &models.AuthResponse{
		Token:  token,
		UserId: user.Id,
	}, nil
}
