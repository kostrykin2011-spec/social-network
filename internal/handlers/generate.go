package handlers

import (
	"bufio"
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"social-network/pkg/models"
	"social-network/pkg/service"
	"strings"

	"github.com/google/uuid"
)

type GenerateHandler interface {
	GenerateData(w http.ResponseWriter, r *http.Request)
}

type generateHandler struct {
	authService       service.AuthService
	friendShipService service.FriendShipService
	postService       service.PostService
}

func InitGenerateHandler(authService service.AuthService, friendShipService service.FriendShipService, postService service.PostService) GenerateHandler {
	return &generateHandler{
		authService:       authService,
		friendShipService: friendShipService,
		postService:       postService,
	}
}

// Генерация пользователей и постов
func (handler *generateHandler) GenerateData(w http.ResponseWriter, r *http.Request) {
	//maxPostCount := 70
	//maxFriendCount := 20
	userIds, err := handler.generateUsers(r.Context())
	if err == nil {
		log.Fatalf(err.Error())
		return
	}

	// posts, _ := handler.getPostsFromFile()
	// allCountPosts := len(posts)
	// countCreatedPosts := 0

	// for _, userId := range userIds {
	// 	for i := 1; i <= maxPostCount; i++ {
	// 		countCreatedPosts++
	// 		randIndex := rand.Intn(allCountPosts)
	// 		err := handler.createPostByUserId(userId, posts[randIndex])
	// 		if err != nil {
	// 		}
	// 	}
	// }

	// // Добавляем каждому пользователю друзей
	// countUserIds := len(userIds)
	// for _, userId := range userIds {
	// 	j := 1
	// 	for {
	// 		randIndex := rand.Intn(countUserIds)
	// 		friendId := userIds[randIndex]
	// 		// Добавляем друга
	// 		handler.friendShipService.AddFiend(userId, friendId)
	// 		j++
	// 		if j > maxFriendCount {
	// 			break
	// 		}
	// 	}
	// }

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Генерация пользователей и постов успешно завершена"})
	json.NewEncoder(w).Encode(userIds)
}

func (handler *generateHandler) generateUsers(ctx context.Context) ([]uuid.UUID, error) {
	filePath := os.Getenv("CSV_PEOPLE_PATH")
	if filePath == "" {
		return nil, fmt.Errorf("Файл people.csv не найден")
	}

	file, err := os.Open(filePath)
	if err != nil {
		fmt.Println("Ошибка при открытии файла:", err)
	}

	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		fmt.Println("Ошибка при чтении CSV:", err)
		return nil, err
	}

	var userIds []uuid.UUID

	count := 0
	allCount := 0

	for _, record := range records {
		fio := record[0]
		parts := strings.Split(fio, " ")
		request := models.RegisterRequest{
			FirstName: parts[0],
			LastName:  parts[1],
			Birthdate: record[1],
			City:      record[2],
			Gender:    "Man",
			Biography: "Информация о пользователе....",
			Password:  "Секретная строка",
		}
		go func() {
			profile, _ := handler.authService.UserRegister(ctx, &request)
			userIds = append(userIds, profile.UserId)
			count++
			if count > 1000 {
				allCount += count
				fmt.Println("Обработано:", allCount)
				count = 0
			}
		}()
	}

	return userIds, nil
}

// Получение постов из файла
func (handler *generateHandler) getPostsFromFile() ([]*models.CreatePostRequest, error) {
	filePath := os.Getenv("CSV_POSTS_PATH")
	if filePath == "" {
		return nil, fmt.Errorf("Файл people.csv не найден")
	}

	file, err := os.Open("posts.txt")
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var posts []*models.CreatePostRequest
	scanner := bufio.NewScanner(file)

	var currentParagraph strings.Builder
	paragraphID := 1

	for scanner.Scan() {
		line := scanner.Text()
		posts = append(posts, &models.CreatePostRequest{
			Content: line,
		})
		paragraphID++
		currentParagraph.Reset()
	}

	if currentParagraph.Len() > 0 {
		content := strings.TrimSpace(currentParagraph.String())
		posts = append(posts, &models.CreatePostRequest{
			Content: content,
		})
	}

	return posts, nil
}

// Создание поста пользователем
func (handler *generateHandler) createPostByUserId(ctx context.Context, userId uuid.UUID, postRequest *models.CreatePostRequest) error {
	postRequest.Title = userId.String()
	err := handler.postService.AddPost(ctx, userId, postRequest)
	if err != nil {
		return err
	}

	return nil
}
