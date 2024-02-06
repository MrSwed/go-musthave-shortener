package repository

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type DBStorage interface {
	Ping() error
}

type DBStorageRepo struct {
	db *pgxpool.Pool
}

func NewDBStorageRepository(db *pgxpool.Pool) *DBStorageRepo {
	return &DBStorageRepo{
		db: db,
	}
}

func (r *DBStorageRepo) Ping() error {
	if r.db == nil {
		return fmt.Errorf("no DB connected")
	}
	return r.db.Ping(context.Background())
}
