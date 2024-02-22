package repository

import (
	"context"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	"github.com/jmoiron/sqlx"
)

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-shortener/internal/app/repository" Repository

type DataStorage interface {
	GetFromShort(ctx context.Context, k string) (string, error)
	GetFromURL(ctx context.Context, url string) (string, error)
	NewShort(ctx context.Context, url string) (newURL string, err error)
	GetAll(ctx context.Context) (Store, error)
	RestoreAll(Store) error
	NewShortBatch(context.Context, []domain.ShortBatchInputItem, string) ([]domain.ShortBatchResultItem, error)
	Ping(ctx context.Context) error
	GetUser(ctx context.Context, id string) (domain.UserInfo, error)
	NewUser(ctx context.Context) (string, error)
	GetAllByUser(ctx context.Context, userID, prefix string) ([]domain.StorageItem, error)
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
	DB          *sqlx.DB
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
