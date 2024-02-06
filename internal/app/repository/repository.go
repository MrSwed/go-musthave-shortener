package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-shortener/internal/app/repository" Repositories

type Repositories interface {
	MemStorage
	FileStorage
	DBStorage
}

type Storage struct {
	MemStorage
	FileStorage
	DBStorage
}

type Config struct {
	StorageFile string
	DB          *pgxpool.Pool
}

func NewRepositories(c Config) Storage {
	return Storage{
		MemStorage:  NewMemRepository(),
		FileStorage: NewFileStorage(c.StorageFile),
		DBStorage:   NewDBStorageRepository(c.DB),
	}
}
