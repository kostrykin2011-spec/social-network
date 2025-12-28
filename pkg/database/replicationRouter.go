package database

import (
	"context"
	"database/sql"
	"sync/atomic"
)

type ReplicationRouter struct {
	master     *sql.DB
	replicates []*sql.DB
	counter    uint64
}

func InitReplicationRouter(master *sql.DB, replicates ...*sql.DB) *ReplicationRouter {
	return &ReplicationRouter{
		master:     master,
		replicates: replicates,
	}
}

func (rep *ReplicationRouter) GetConnection(ctx context.Context, operation string) (*sql.DB, error) {
	switch operation {
	case "write":
		return rep.master, nil
	case "read":
		if len(rep.replicates) == 0 {
			return rep.master, nil
		}

		ids := atomic.AddUint64(&rep.counter, 1) % uint64(len(rep.replicates))
		return rep.replicates[ids], nil
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

func (rep *ReplicationRouter) GetOperationType(ctx context.Context) string {
	if operation, ok := ctx.Value("operation").(string); ok {
		return operation
	}

	return "read"
}

// По переданному в контекст значения "read/write", вычисляем БД
func (replicator *ReplicationRouter) GetDatabase(ctx context.Context) (*sql.DB, error) {
	operation := replicator.GetOperationType(ctx)

	return replicator.GetConnection(ctx, operation)
}
