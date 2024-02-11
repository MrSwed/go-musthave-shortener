package repository

import "github.com/MrSwed/go-musthave-shortener/internal/app/config"

func NewRepository(c *config.Config) MemStorage {
	return NewMemRepository(c)
}
