package services

import (
	"context"
	"counter-service/internal/repository"
	logger "counter-service/pkg"
)

type CounterService struct {
	dbRepo    *repository.CounterDBRepository
	redisRepo *repository.CounterCacheRepository
	logger    *logger.Logger
}

func InitCounterService(dbRepo *repository.CounterDBRepository, redisRepo *repository.CounterCacheRepository, logger *logger.Logger) *CounterService {
	return &CounterService{
		dbRepo:    dbRepo,
		redisRepo: redisRepo,
		logger:    logger,
	}
}

// Возвращаем количество непрочитанных сообщений по (пользователю и диалогу)
func (service *CounterService) GetCount(ctx context.Context, recipientId, userId string) (int64, error) {
	countByRedis, err := service.redisRepo.Get(ctx, userId, recipientId)
	if err == nil {
		return countByRedis, err
	}

	countByDB, err := service.dbRepo.Get(ctx, userId, recipientId)
	if err != nil {
		return 0, err
	}

	service.redisRepo.Set(ctx, userId, recipientId, countByDB)

	return countByDB, nil
}

// Сбрасываем счетчик непрочитанных сообщений, когда пользователь открыл чат (диалог)
func (service *CounterService) Reset(ctx context.Context, RecipientId, userId string) error {
	err := service.redisRepo.Reset(ctx, userId, RecipientId)
	if err != nil {
		return err
	}

	return service.dbRepo.Reset(ctx, userId, RecipientId)
}
