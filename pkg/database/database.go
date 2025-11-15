package database

import (
	"database/sql"
	"fmt"
	_ "log"
	_ "time"

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

	err = createTables(db)

	if err != nil {
		return nil, fmt.Errorf("Не удалось создать таблицу: %w", err)
	}

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
			first_name VARCHAR(100) NOT NULL,
			last_name VARCHAR(100) NOT NULL,
			birth_date DATE NOT NULL,
			gender VARCHAR(20) NOT NULL,
			biography TEXT,
			city VARCHAR(100),
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		);`

	_, err := db.Exec(query)

	return err
}
