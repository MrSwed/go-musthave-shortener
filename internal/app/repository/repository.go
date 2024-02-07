package repository

import (
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-shortener/internal/app/repository" Repository

type DataStorage interface {
	GetFromShort(k string) (string, error)
	NewShort(url string) (newURL string, err error)
	GetAll() (Store, error)
	RestoreAll(Store) error
}

type Repository interface {
	DataStorage
	FileStorage
}

type Storage struct {
	DataStorage
	FileStorage
}

type Config struct {
	StorageFile string
	DB          *pgxpool.Pool
}

func NewRepository(c Config) (s Storage) {
	if c.DB != nil {
		s = Storage{
			FileStorage: NewFileStorage(c.StorageFile),
			DataStorage: NewDBStorageRepository(c.DB),
		}
	} else {
		s = Storage{
			FileStorage: NewFileStorage(c.StorageFile),
			DataStorage: NewMemRepository(),
		}
	}
	return s
}
