package handlers

import (
	"bufio"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"social-network/pkg/database"
	"social-network/pkg/models"
	"social-network/pkg/service"
	"social-network/pkg/utils"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	BatchSize     = 5000
	ChannelBuffer = 10000
)

type GenerateHandler interface {
	GenerateUsers(w http.ResponseWriter, r *http.Request)
}

type RawRecord struct {
	LineNumber int
	FullName   string
	BirthDate  string
	City       string
}

type ParsedRecord struct {
	User    models.User
	Profile models.Profile
}

type ParsedResult struct {
	userId uuid.UUID
}

type generateHandler struct {
	routerDB          *database.DBRouter
	authService       service.AuthService
	friendShipService service.FriendShipService
	postService       service.PostService
}

func InitGenerateHandler(routerDB *database.DBRouter, authService service.AuthService, friendShipService service.FriendShipService, postService service.PostService) GenerateHandler {
	return &generateHandler{
		routerDB:          routerDB,
		authService:       authService,
		friendShipService: friendShipService,
		postService:       postService,
	}
}

// Генерация пользователей
func (handler *generateHandler) GenerateUsers(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)

	go func() {
		defer cancel()
		handler.generateUsers(ctx)
	}()

	w.WriteHeader(http.StatusAccepted)
	json.NewEncoder(w).Encode(map[string]string{"message": "Генерация пользователей запущена"})
}

func (handler *generateHandler) generateUsers(ctx context.Context) {
	ctx = database.WithMaster(ctx)
	db, err := handler.routerDB.GetDatabase(ctx)
	if err != nil {
		return
	}

	filePath := os.Getenv("CSV_PEOPLE_PATH")
	if filePath == "" {
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		return
	}
	defer file.Close()

	rawChan := make(chan RawRecord, 10000)
	parsedChan := make(chan ParsedRecord, 10000)
	resultChan := make(chan ParsedResult, 10000)

	var wgParseFile sync.WaitGroup
	var wgDB sync.WaitGroup
	var userIds []uuid.UUID
	var userIdsMu sync.Mutex
	WorkersParse := runtime.NumCPU()
	WorkersDB := runtime.NumCPU()

	resultDone := make(chan struct{})
	go func() {
		for result := range resultChan {
			userIdsMu.Lock()
			userIds = append(userIds, result.userId)
			userIdsMu.Unlock()
		}
		close(resultDone)
	}()

	for i := 0; i < WorkersParse; i++ {
		wgParseFile.Add(1)
		go handler.parseWorker(i, rawChan, parsedChan, resultChan, &wgParseFile)
	}

	for i := 0; i < WorkersDB; i++ {
		wgDB.Add(1)
		go handler.dbWorker(i, parsedChan, &wgDB, BatchSize, db)
	}

	reader := csv.NewReader(file)
	reader.FieldsPerRecord = 3
	reader.ReuseRecord = true

	lineNum := 1
	recordsRead := 0

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			lineNum++
			continue
		}

		rawRecord := RawRecord{
			LineNumber: lineNum,
			FullName:   strings.TrimSpace(record[0]),
			BirthDate:  strings.TrimSpace(record[1]),
			City:       strings.TrimSpace(record[2]),
		}

		select {
		case rawChan <- rawRecord:
			recordsRead++
		case <-ctx.Done():
			close(rawChan)
			return
		}

		lineNum++
	}

	close(rawChan)
	wgParseFile.Wait()
	close(parsedChan)
	wgDB.Wait()
	close(resultChan)
	<-resultDone

	// userIdsMu.Lock()
	// defer userIdsMu.Unlock()

	// randomIndex := rand.N(len(userIds))

	// currentUserId := userIds[randomIndex]
	// log.Println("Пользователь, получающий события от друзей на websocket-сервер: " + currentUserId.String())

	// var friendIds []uuid.UUID

	// // Добавляем текущему пользователю - 3 друзей
	// countUserIds := len(userIds)

	// for i := 0; i < 3; i++ {
	// 	randomIndex = rand.N(countUserIds)
	// 	friendId := userIds[randomIndex]
	// 	if friendId == currentUserId {
	// 		continue
	// 	}

	// 	log.Printf("Автор постов #%v: %v", strconv.Itoa(i), friendId.String())

	// 	friendIds = append(friendIds, friendId)
	// 	err := handler.friendShipService.AddFiend(ctx, currentUserId, friendId)
	// 	if err != nil {
	// 	}
	// }
	// if len(friendIds) <= 0 {
	// 	return
	// }

	// posts, err := handler.getPostsFromFile()
	// if err != nil {
	// 	log.Println(err.Error())
	// 	return
	// }
	// stopWorkersUser := handler.StartUsersInBackground(friendIds, posts, len(posts))
	// c := make(chan os.Signal, 1)
	// signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	// <-c
	// stopWorkersUser()
}

// Друзья будут публиковать посты в разное время (от 1 до 20 сек) бесконечно (до отмены)
func (handler *generateHandler) StartUsersInBackground(userIDs []uuid.UUID, posts []*models.CreatePostRequest, postCount int) context.CancelFunc {
	ctx, cancel := context.WithCancel(context.Background())
	for _, userID := range userIDs {
		go func(ctx context.Context, id uuid.UUID, posts []*models.CreatePostRequest, postCount int, handler *generateHandler) {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					randomPostIndex := rand.N(postCount)
					err := handler.createPostByUserId(ctx, userID, posts[randomPostIndex])
					if err != nil {
						log.Printf(err.Error())
					}

					waitSeconds := 1 + rand.IntN(20)
					select {
					case <-time.After(time.Duration(waitSeconds) * time.Second):
					case <-ctx.Done():
						return
					}
				}
			}
		}(ctx, userID, posts, postCount, handler)
	}

	return cancel
}

func (handler *generateHandler) parseWorker(id int, rawChan <-chan RawRecord, parsedChan chan<- ParsedRecord, resultChan chan<- ParsedResult, wgParseFile *sync.WaitGroup) {
	defer func() {
		wgParseFile.Done()
	}()

	pass, _ := utils.HashPassword("Секретная строка")

	for raw := range rawChan {
		parts := strings.Split(raw.FullName, " ")
		if len(parts) < 2 {
			continue
		}

		birthDate, err := time.Parse("2006-01-02", raw.BirthDate)
		if err != nil {
			continue
		}

		userId := uuid.New()
		resultChan <- ParsedResult{
			userId: userId,
		}
		user := models.User{
			Id:       userId,
			Password: pass,
		}

		profile := models.Profile{
			Id:        uuid.New(),
			UserId:    userId,
			FirstName: parts[0],
			LastName:  parts[1],
			Birthdate: birthDate,
			Gender:    "Man",
			Biography: "Информация о себе...",
			City:      raw.City,
		}

		// Отправляем в parsedChan
		parsedChan <- ParsedRecord{
			User:    user,
			Profile: profile,
		}
	}
}

func (handler *generateHandler) dbWorker(id int, parsedChan <-chan ParsedRecord, wg *sync.WaitGroup, batchSize int, db *sql.DB) {
	defer func() {
		wg.Done()
	}()

	var usersBatch []models.User
	var profilesBatch []models.Profile
	var totalProcessed int

	userStmt, err := db.Prepare(`
		INSERT INTO users (id, password) 
		VALUES ($1, $2) 
		RETURNING id
	`)
	if err != nil {
		//log.Fatal(err)
	}
	defer userStmt.Close()

	profileStmt, err := db.Prepare(`
		INSERT INTO profiles (id, user_id, first_name, last_name, birth_date, gender, biography, city) VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING created_at`)
	if err != nil {
		//log.Fatal(err)
	}
	defer profileStmt.Close()

	for record := range parsedChan {
		totalProcessed++
		usersBatch = append(usersBatch, record.User)
		profilesBatch = append(profilesBatch, record.Profile)

		if len(usersBatch) >= batchSize {
			handler.flushBatchWithStmt(db, userStmt, profileStmt, usersBatch, profilesBatch)
			usersBatch = usersBatch[:0]
			profilesBatch = profilesBatch[:0]
		}
	}

	if len(usersBatch) > 0 {
		handler.flushBatchWithStmt(db, userStmt, profileStmt, usersBatch, profilesBatch)
	}
}

func (handler *generateHandler) flushBatchWithStmt(db *sql.DB, userStmt, profileStmt *sql.Stmt, users []models.User, profiles []models.Profile) {
	tx, err := db.Begin()
	if err != nil {
	}
	defer tx.Rollback()

	txUserStmt := tx.Stmt(userStmt)
	txProfileStmt := tx.Stmt(profileStmt)
	defer txUserStmt.Close()
	defer txProfileStmt.Close()

	for i, user := range users {
		_, err := txUserStmt.Exec(user.Id, user.Password)
		if err != nil {
			log.Printf("Ошибка вставки user %d: %v", i, err)
			continue
		}
	}

	for i, profile := range profiles {
		_, err := txProfileStmt.Exec(
			profile.Id,
			profile.UserId,
			profile.FirstName,
			profile.LastName,
			profile.Birthdate,
			profile.Gender,
			profile.Biography,
			profile.City,
		)
		if err != nil {
			log.Printf("Ошибка вставки profile %d: %v", i, err)
			continue
		}
	}
	if err = tx.Commit(); err != nil {
		//log.Fatal(err)
	}

}

// Получение постов из файла
func (handler *generateHandler) getPostsFromFile() ([]*models.CreatePostRequest, error) {
	filePath := os.Getenv("CSV_POSTS_PATH")
	if filePath == "" {
		return nil, fmt.Errorf("Файл post.txt не найден")
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
