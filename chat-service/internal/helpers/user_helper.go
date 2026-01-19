package helpers

import (
	"strings"

	"github.com/google/uuid"
)

// Особая сортировка для правильного распределения данных при шардинге
func OrderUUIDs(a, b uuid.UUID) (uuid.UUID, uuid.UUID) {
	if strings.Compare(a.String(), b.String()) < 0 {
		return a, b
	}
	return b, a
}
