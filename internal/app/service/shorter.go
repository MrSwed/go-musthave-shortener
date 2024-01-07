package service

import (
	"errors"
	"fmt"
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"

	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
)

type Shorter interface {
	NewShort(url string) (string, error)
	GetFromShort(k string) (string, error)
}

type ShorterService struct {
	r repository.MemStorage
	c *config.Config
}

func NewShorterService(r repository.MemStorage, c *config.Config) *ShorterService {
	return &ShorterService{r: r, c: c}
}

func (s *ShorterService) NewShort(url string) (newURL string, err error) {
	var sk config.ShortKey
	for {
		sk = helper.NewRandShorter().RandStringBytes()
		if _, chErr := s.r.GetFromShort(sk); chErr != nil && errors.Is(chErr, myErr.ErrNotExist) {
			err = s.r.SaveShort(sk, url)
			break
		}
	}
	newURL = fmt.Sprintf("%s%s/%s", s.c.Scheme, s.c.BaseURL, sk)
	return
}

func (s *ShorterService) GetFromShort(k string) (v string, err error) {
	v, err = s.r.GetFromShort(config.ShortKey([]byte(k)))
	return
}
