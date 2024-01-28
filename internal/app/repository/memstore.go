package repository

import (
	"fmt"
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"
)

type MemStorage interface {
	GetFromShort(k config.ShortKey) (string, error)
	NewShort(url string) (newURL string, err error)
}

type MemStorageRepository struct {
	c  *config.Config
	db map[config.ShortKey]string
	mg sync.RWMutex
}

func NewMemRepository(c *config.Config) *MemStorageRepository {
	return &MemStorageRepository{
		db: make(map[config.ShortKey]string),
		c:  c,
	}
}

func (r *MemStorageRepository) NewShort(url string) (newURL string, err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	for newShort := helper.NewRandShorter().RandStringBytes(); ; {
		if _, exist := r.db[newShort]; !exist {
			r.db[newShort] = url
			newURL = fmt.Sprintf("%s%s/%s", r.c.Scheme, r.c.BaseURL, newShort)
			return
		}
	}
}

func (r *MemStorageRepository) GetFromShort(k config.ShortKey) (v string, err error) {
	var ok bool
	r.mg.RLock()
	defer r.mg.RUnlock()
	if v, ok = r.db[k]; !ok {
		err = myErr.ErrNotExist
	}
	return
}
