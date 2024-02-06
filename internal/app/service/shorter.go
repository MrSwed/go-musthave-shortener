package service

import (
	"fmt"
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
)

type Shorter interface {
	NewShort(url string) (string, error)
	GetFromShort(k string) (string, error)
	CheckDB() error
}

type ShorterService struct {
	r repository.Repositories
	c *config.Config
}

func NewShorterService(r repository.Repositories, c *config.Config) ShorterService {
	return ShorterService{r: r, c: c}
}

func (s ShorterService) NewShort(url string) (newURL string, err error) {
	var newShort string
	if newShort, err = s.r.NewShort(url); err == nil {
		newURL = fmt.Sprintf("%s%s/%s", s.c.Scheme, s.c.BaseURL, newShort)
	}

	return
}

func (s ShorterService) GetFromShort(k string) (v string, err error) {
	v, err = s.r.GetFromShort(k)
	return
}

func (s ShorterService) CheckDB() error {
	return s.r.Ping()
}
