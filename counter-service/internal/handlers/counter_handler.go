package handlers

import (
	"context"
	"counter-service/internal/models"
	"counter-service/internal/services"
	logger "counter-service/pkg"
)

type CounterHandler struct {
	counterService *services.CounterService
	log            *logger.Logger
}

func InitCounterHandler(counterService *services.CounterService, log *logger.Logger) *CounterHandler {
	return &CounterHandler{
		counterService: counterService,
		log:            log,
	}
}

func (handler *CounterHandler) GetCount(ctx context.Context, unreadCountRequest *models.UnreadCountRequest) (*models.UnreadCountResponse, error) {
	count, err := handler.counterService.GetCount(ctx, unreadCountRequest.RecipientId, unreadCountRequest.UserId)
	if err != nil {
		return nil, err
	}

	return &models.UnreadCountResponse{
		UserId:      unreadCountRequest.UserId,
		RecipientId: unreadCountRequest.RecipientId,
		Count:       count,
	}, nil
}

func (handler *CounterHandler) Reset(ctx context.Context, countResetRequest *models.CountResetRequest) (*models.CountResetResponse, error) {
	err := handler.counterService.Reset(ctx, countResetRequest.RecipientId, countResetRequest.UserId)
	if err != nil {
		return nil, err
	}

	return &models.CountResetResponse{
		UserId:      countResetRequest.UserId,
		RecipientId: countResetRequest.RecipientId,
		Success:     true,
	}, nil
}
