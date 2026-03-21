package database

import (
	"context"
	"database/sql"
)

type DBRouter struct {
	master *sql.DB
	slave  *sql.DB
}

func InitDBRouter(master *sql.DB, slave *sql.DB) *DBRouter {
	return &DBRouter{
		master: master,
		slave:  slave,
	}
}

func (rep *DBRouter) GetConnection(ctx context.Context, operation string) (*sql.DB, error) {
	switch operation {
	case "write":
		return rep.master, nil
	case "read":
		return rep.slave, nil
	default:
		return rep.master, nil
	}
}

func WithMaster(ctx context.Context) context.Context {
	return context.WithValue(ctx, "operation", "write")
}

func WithReplica(ctx context.Context) context.Context {
	return context.WithValue(ctx, "operation", "read")
}

func (rep *DBRouter) GetOperationType(ctx context.Context) string {
	if operation, ok := ctx.Value("operation").(string); ok {
		return operation
	}

	return "read"
}

func (replicator *DBRouter) GetDatabase(ctx context.Context) (*sql.DB, error) {
	operation := replicator.GetOperationType(ctx)

	return replicator.GetConnection(ctx, operation)
}
