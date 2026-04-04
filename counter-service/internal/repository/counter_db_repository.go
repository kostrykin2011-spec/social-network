package repository

import (
	"context"
	"database/sql"
)

type CounterDBRepository struct {
	db *sql.DB
}

// Инициализация репозитория счетчиков для БД
func InitCounterDBRepository(db *sql.DB) *CounterDBRepository {
	return &CounterDBRepository{
		db: db,
	}
}

// Получение количества непрочитанных сообщений
func (repository *CounterDBRepository) Get(ctx context.Context, userId, recipientId string) (int64, error) {
	var result int64
	query := `SELECT unread_count FROM counters WHERE user_id = $1 AND recipient_id = $2`
	err := repository.db.QueryRowContext(ctx, query, userId, recipientId).Scan(&result)
	if err == sql.ErrNoRows {
		return 0, nil
	}

	return result, nil
}

// Увеличение на 1
func (repository *CounterDBRepository) Increment(ctx context.Context, sagaId, userId, recipientId string) error {
	tx, err := repository.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `INSERT INTO processed_events (saga_id) VALUES ($1)`, sagaId)
	if err != nil {
		return nil
	}

	query := `
        INSERT INTO counters (user_id, recipient_id, unread_count, updated_at)
        VALUES ($1, $2, 1, NOW())
        ON CONFLICT (user_id, recipient_id)
        DO UPDATE SET 
            unread_count = counters.unread_count + 1,
            updated_at = NOW()
    `
	_, err = tx.ExecContext(ctx, query, userId, recipientId)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// Уменьшение на 1
func (repository *CounterDBRepository) Decrement(ctx context.Context, userId, recipientId string) error {
	query := `
        UPDATE counters 
        SET unread_count = GREATEST(unread_count - 1, 0),
            updated_at = NOW()
        WHERE user_id = $1 AND recipient_id = $2
    `
	_, err := repository.db.ExecContext(ctx, query, userId, recipientId)

	return err
}

// Уменьшение на 1
func (repository *CounterDBRepository) Reset(ctx context.Context, userId, recipientId string) error {
	query := `
        UPDATE counters 
        SET unread_count = 0,
            updated_at = NOW()
        WHERE user_id = $1 AND recipient_id = $2
    `
	_, err := repository.db.ExecContext(ctx, query, userId, recipientId)

	return err
}
