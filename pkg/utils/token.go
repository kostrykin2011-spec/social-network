package utils

import (
	"sync"

	"github.com/google/uuid"
)

var (
	mutex  = &sync.RWMutex{}
	tokens = make(map[string]string)
)

func GenerateToken() string {
	return uuid.New().String()
}

func SaveToken(token, userId string) {
	mutex.Lock()
	defer mutex.Unlock()

	tokens[token] = userId
}

func ValidateToken(token string) (string, bool) {
	mutex.RLock()
	defer mutex.RUnlock()

	userId, exists := tokens[token]

	return userId, exists
}

func DeleteToken(token string) {
	mutex.Lock()
	defer mutex.Unlock()

	delete(tokens, token)
}
