package repository

import "github.com/MrSwed/go-musthave-shortener/internal/app/config"

//go:generate  mockgen -destination=../mock/repository/repository.go -package=mock "github.com/MrSwed/go-musthave-shortener/internal/app/repository" Repository

type Repository interface {
	MemStorage
	FileStorage
}

type Storage struct {
	MemStorage
	FileStorage
}

func NewRepository(c *config.Config) Storage {
	return Storage{
		MemStorage:  NewMemRepository(c),
		FileStorage: NewFileStorage(c.FileStoragePath),
	}
}
