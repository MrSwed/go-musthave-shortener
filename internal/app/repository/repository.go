package repository

import "github.com/MrSwed/go-musthave-shortener/internal/app/config"

type Repository struct {
	MemStorage
	FileStorage
}

func NewRepository(c *config.Config) Repository {
	return Repository{
		MemStorage:  NewMemRepository(c),
		FileStorage: NewFileStorage(c.FileStoragePath),
	}
}
