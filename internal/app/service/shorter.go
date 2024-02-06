package service

import (
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
)

type Shorter interface {
	NewShort(url string) (string, error)
	GetFromShort(k string) (string, error)
	CheckDB() error
}

type ShorterService struct {
	r repository.Repository
}

func NewShorterService(r repository.Repository) ShorterService {
	return ShorterService{r: r}
}

func (s ShorterService) NewShort(url string) (newURL string, err error) {
	return s.r.NewShort(url)
}

func (s ShorterService) GetFromShort(k string) (v string, err error) {
	v, err = s.r.GetFromShort(k)
	return
}

func (s ShorterService) CheckDB() error {
	return s.r.Ping()
}
