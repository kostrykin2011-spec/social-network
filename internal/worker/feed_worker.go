package worker

import (
	"context"
	"fmt"
	"log"
	"social-network/internal/queue"
	"social-network/pkg/database"
	"social-network/pkg/repository"
	"social-network/pkg/service"
)

type FeedWorker struct {
	postRepository repository.PostRepository
	feedService    service.FeedService
	rabbitMQ       *queue.RabbitMQClient
	queueName      string
}

func NewFeedWorker(
	postRepository repository.PostRepository,
	feedService service.FeedService,
	queue *queue.RabbitMQClient,
	queueName string,
) *FeedWorker {
	return &FeedWorker{
		postRepository: postRepository,
		feedService:    feedService,
		rabbitMQ:       queue,
		queueName:      queueName,
	}
}

func (w *FeedWorker) Start(ctx context.Context) error {
	return w.rabbitMQ.ConsumeFeedTasks(ctx, w.queueName, w.processTask)
}

// Получаем задачи из очереди
func (w *FeedWorker) processTask(ctx context.Context, task *queue.FeedTask) error {
	ctx = database.WithMaster(ctx)
	// 1. Получаем информацию о посте из БД
	post, err := w.postRepository.GetById(ctx, task.PostID)
	if err != nil {
		return err
	}

	if post == nil {
		fmt.Println("Пост не найден")
	}

	// Добавляем пост в кеш и публикуем события для веб-сокет в очереди RabbitMQ
	err = w.feedService.AddPostToFeed(ctx, post.UserId, post)
	if err != nil {
		log.Printf("Не удалось добавить пост: %v", err)
	}

	return nil
}
