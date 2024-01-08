package service

import (
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
)

type Service struct {
	Shorter
}

func NewService(r repository.MemStorage, c *config.Config) Service {
	return Service{Shorter: NewShorterService(r, c)}
}
