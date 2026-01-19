package helpers

import (
	"math/rand"
	"sync"
	"time"

	"github.com/google/uuid"
)

type uuidCache struct {
	mu      sync.RWMutex
	items   map[int]uuid.UUID // индекс -> UUID
	indices []int             // все существующие индексы
	nextID  int               // следующий свободный индекс
}

var (
	instance *uuidCache
	once     sync.Once
)

// Instance возвращает глобальный экземпляр кеша (синглтон)
func Instance() *uuidCache {
	once.Do(func() {
		rand.Seed(time.Now().UnixNano())
		instance = &uuidCache{
			items:   make(map[int]uuid.UUID),
			indices: make([]int, 0),
			nextID:  0,
		}
	})
	return instance
}

// Add добавляет UUID и возвращает его индекс
func (c *uuidCache) Add(id uuid.UUID) int {
	c.mu.Lock()
	defer c.mu.Unlock()

	index := c.nextID
	c.items[index] = id
	c.indices = append(c.indices, index)
	c.nextID++

	return index
}

// GetByIndex возвращает UUID по индексу
func (c *uuidCache) GetByIndex(index int) (uuid.UUID, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	id, exists := c.items[index]
	return id, exists
}

// Remove удаляет UUID по индексу
func (c *uuidCache) Remove(index int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.items[index]; !exists {
		return false
	}

	delete(c.items, index)

	// Находим и удаляем индекс из слайса
	for i, idx := range c.indices {
		if idx == index {
			// Быстрое удаление (меняем с последним)
			c.indices[i] = c.indices[len(c.indices)-1]
			c.indices = c.indices[:len(c.indices)-1]
			break
		}
	}

	return true
}

// GetRandom возвращает случайный UUID за O(1)
func (c *uuidCache) GetRandom() (uuid.UUID, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.indices) == 0 {
		return uuid.Nil, false
	}

	randomIndex := c.indices[rand.Intn(len(c.indices))]
	return c.items[randomIndex], true
}

// GetRandomN возвращает n случайных UUID
func (c *uuidCache) GetRandomN(n int) []uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if len(c.indices) == 0 || n <= 0 {
		return []uuid.UUID{}
	}

	if n > len(c.indices) {
		n = len(c.indices)
	}

	result := make([]uuid.UUID, 0, n)

	if n < len(c.indices)/2 {
		// Для небольшого количества - случайные неповторяющиеся выборки
		used := make(map[int]bool, n)
		for len(result) < n {
			idx := c.indices[rand.Intn(len(c.indices))]
			if !used[idx] {
				used[idx] = true
				result = append(result, c.items[idx])
			}
		}
	} else {
		// Для большого количества - перемешиваем и берем первые n
		shuffled := make([]int, len(c.indices))
		copy(shuffled, c.indices)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})

		for i := 0; i < n; i++ {
			result = append(result, c.items[shuffled[i]])
		}
	}

	return result
}

// Size возвращает количество элементов
func (c *uuidCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.indices)
}

// GetAll возвращает все UUID
func (c *uuidCache) GetAll() []uuid.UUID {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]uuid.UUID, 0, len(c.indices))
	for _, idx := range c.indices {
		result = append(result, c.items[idx])
	}
	return result
}

// Clear очищает кеш
func (c *uuidCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.items = make(map[int]uuid.UUID)
	c.indices = make([]int, 0)
	c.nextID = 0
}
