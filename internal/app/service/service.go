package service

import (
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
)

type Service struct {
	Shorter
}

func NewService(r repository.MemStorage) Service {
	return Service{Shorter: NewShorterService(r)}
}
