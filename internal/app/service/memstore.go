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
}

func NewMemStorage(r repository.MemStorage) *ShorterService {
	return &ShorterService{r: r}
}

func (s *ShorterService) NewShort(url string) (newUrl string, err error) {
	var sk config.ShortKey
	for {
		sk = helper.NewRandShorter().RandStringBytes()
		if _, chErr := s.r.GetFromShort(sk); chErr != nil && errors.Is(chErr, myErr.ErrNotExist) {
			err = s.r.SaveShort(sk, url)
			break
		}
	}
	newUrl = fmt.Sprintf("%s%s/%s", config.Scheme, config.Address, sk)
	return
}

func (s *ShorterService) GetFromShort(k string) (v string, err error) {
	var vS config.ShortKey
	for i := 0; i < config.ShortLen; i++ {
		vS[i] = k[i]
	}
	v, err = s.r.GetFromShort(vS)
	return
}
