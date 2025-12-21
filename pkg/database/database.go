package database

import (
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
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	return db, nil
}

func createTables(db *sql.DB) error {
	query := `
		CREATE TABLE IF NOT EXISTS users (
			id UUID PRIMARY KEY NOT NULL,
			password VARCHAR(255) NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS profiles (
			id UUID PRIMARY KEY NOT NULL,
			user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			last_name VARCHAR(100) NOT NULL,
			first_name VARCHAR(100) NOT NULL,
			birth_date DATE NOT NULL,
			gender VARCHAR(20) NOT NULL,
			biography TEXT,
			city VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);
		CREATE TABLE IF NOT EXISTS friendships (
        	id UUID PRIMARY KEY NOT NULL,
        	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        	friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        	status VARCHAR(20) DEFAULT 'is_friend',
        	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
        	UNIQUE(user_id, friend_id)
		);
		CREATE TABLE IF NOT EXISTS posts (
        	id UUID PRIMARY KEY NOT NULL,
        	user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
        	title VARCHAR(255) NOT NULL,
        	content TEXT NOT NULL,
			is_public BOOLEAN DEFAULT false,
        	created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

	_, err := db.Exec(query)

	return err
}

func CreateIndexes(db *sql.DB) error {
	query := `
		CREATE INDEX IF NOT EXISTS idx_profiles_fullname_btree ON profiles (
			last_name varchar_pattern_ops, 
			first_name varchar_pattern_ops
		);`

	_, err := db.Exec(query)

	return err
}
