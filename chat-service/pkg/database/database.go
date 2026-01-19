package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

// InitDatabase - инициализация БД
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
	db.SetConnMaxLifetime(10 * time.Minute)

	// Проверка соединения
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

// createTables - cоздание таблиц
func createTables(db *sql.DB) error {
	query := `
		CREATE EXTENSION IF NOT EXISTS citus;
		SELECT citus_set_coordinator_host('citus-worker-1'); 
		SELECT citus_set_coordinator_host('citus-worker-2');
	 	CREATE TABLE IF NOT EXISTS dialogs (
	  		id UUID NOT NULL,
	  		user1_id UUID NOT NULL,
	  		user2_id UUID NOT NULL,
	  		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id, user1_id),
    		UNIQUE(user1_id, user2_id)
	 	);

	 	CREATE TABLE IF NOT EXISTS messages (
	 		id UUID NOT NULL,
	 		dialog_id UUID NOT NULL,
	 		user1_id UUID NOT NULL,
	 		sender_id UUID NOT NULL,
	 		recipient_id UUID NOT NULL,
	 		content TEXT NOT NULL,
	 		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			PRIMARY KEY (id, user1_id)
	 	);

		SELECT create_distributed_table('dialogs', 'user1_id');
		SELECT create_distributed_table('messages', 'user1_id', colocate_with => 'dialogs');
		ALTER TABLE messages ADD FOREIGN KEY (dialog_id, user1_id) REFERENCES dialogs(id, user1_id);
		`

	_, err := db.Exec(query)

	return err
}

// createTables - cоздание индексов
func CreateIndexes(db *sql.DB) error {
	query := `
	 	CREATE INDEX IF NOT EXISTS idx_dialogs_user1_id ON dialogs (user1_id);
	 	CREATE INDEX IF NOT EXISTS idx_dialogs_user2_id ON dialogs (user2_id);
	 	CREATE INDEX IF NOT EXISTS idx_messages_sender_id ON messages (sender_id);
	 	CREATE INDEX IF NOT EXISTS idx_messages_recipient_id ON messages (recipient_id);
	 	`

	_, err := db.Exec(query)

	return err
}
