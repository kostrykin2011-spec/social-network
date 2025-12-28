package handlers

import (
	"encoding/json"
	"net/http"
	"social-network/pkg/database"
	"strconv"

	"github.com/google/uuid"
)

type TestHandler interface {
	AddRecord(w http.ResponseWriter, r *http.Request)
	GetRecord(w http.ResponseWriter, r *http.Request)
}

type testHandler struct {
	routerDB *database.ReplicationRouter
}

func InitTestHandler(routerDB *database.ReplicationRouter) TestHandler {
	return &testHandler{
		routerDB: routerDB,
	}
}

func (t *testHandler) AddRecord(w http.ResponseWriter, r *http.Request) {
	ctx := database.WithMaster(r.Context())
	db, err := t.routerDB.GetDatabase(ctx)
	id := uuid.New()
	query := `INSERT INTO test (id) VALUES ($1)`

	_ = db.QueryRowContext(ctx, query, &id).Scan()

	var count int
	query = `SELECT COUNT(*) from test`
	err = db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"Количество записей": strconv.Itoa(count)})
}

func (t *testHandler) GetRecord(w http.ResponseWriter, r *http.Request) {
	ctx := database.WithReplica(r.Context())
	db, err := t.routerDB.GetDatabase(ctx)
	var count int
	query := `SELECT COUNT(*) from test`
	err = db.QueryRowContext(ctx, query).Scan(&count)
	if err != nil {
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"Количество записей": strconv.Itoa(count)})
}
