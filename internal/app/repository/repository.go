package repository

import (
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-shortener/internal/app/repository" Repository

type Repository interface {
	MemStorage
	FileStorage
	DBStorage
}

type Storage struct {
	MemStorage
	FileStorage
	DBStorage
}

func NewRepository(c *config.Config, db *pgxpool.Pool) Storage {
	return Storage{
		MemStorage:  NewMemRepository(c),
		FileStorage: NewFileStorage(c.FileStoragePath),
		DBStorage:   NewDBStorageRepository(db),
	}
}
