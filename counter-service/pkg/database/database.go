package database

import (
	"context"
	"database/sql"
	"fmt"
	_ "log"
	"time"

	_ "github.com/lib/pq"
)

func InitDatabase(connStr string) (*sql.DB, error) {
	var err error

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("База данных не доступна: %v", err)
	}
	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("База данных не доступна: %v", err)
	}

	// Создание таблиц
	err = createTables(db)
	if err != nil {
		return nil, fmt.Errorf("Не удалось создать таблицы: %w", err)
	}

	// Создание индексов
	err = CreateIndexes(db)
	if err != nil {
		return nil, fmt.Errorf("Не удалось создать индексы: %w", err)
	}

	// Настройка пула соединений
	db.SetMaxOpenConns(100)
	db.SetMaxIdleConns(50)
	db.SetConnMaxLifetime(1 * time.Minute)

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func createTables(db *sql.DB) error {
	query := `
		CREATE TABLE counters (
    		user_id UUID NOT NULL,
    		recipient_id UUID NOT NULL,
    		unread_count INT DEFAULT 0,
    		updated_at TIMESTAMP DEFAULT NOW(),
   	 	PRIMARY KEY (user_id, recipient_id)
		);
		CREATE TABLE IF NOT EXISTS processed_events (
        	saga_id UUID PRIMARY KEY NOT NULL
		);`

	_, err := db.Exec(query)

	return err
}

func CreateIndexes(db *sql.DB) error {
	query := `
		CREATE INDEX IF NOT EXISTS idx_counters_user_id ON counters(user_id);
		CREATE INDEX IF NOT EXISTS idx_counters_recipient_id ON counters(recipient_id);`

	_, err := db.Exec(query)

	return err
}
