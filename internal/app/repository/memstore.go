package repository

import (
	"fmt"
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"
)

type storeItem struct {
	uuid string
	url  string
}

type Store map[config.ShortKey]storeItem

type MemStorage interface {
	GetFromShort(k config.ShortKey) (string, error)
	NewShort(url string) (newURL string, err error)
	GetAll() Store
	RestoreAll(Store)
}

type MemStorageRepository struct {
	c    *config.Config
	Data Store
	mg   sync.RWMutex
}

func NewMemRepository(c *config.Config) *MemStorageRepository {
	return &MemStorageRepository{
		Data: make(Store),
		c:    c,
	}
}

func (r *MemStorageRepository) NewShort(url string) (newURL string, err error) {
	r.mg.RLock()
	defer r.mg.RUnlock()
	for newShort := helper.NewRandShorter().RandStringBytes(); ; {
		if _, exist := r.Data[newShort]; !exist {
			r.Data[newShort] = storeItem{
				// tmp use int instead uuid
				uuid: fmt.Sprint(len(r.Data) + 1),
				url:  url,
			}
			newURL = fmt.Sprintf("%s%s/%s", r.c.Scheme, r.c.BaseURL, newShort)
			return
		}
	}
}

func (r *MemStorageRepository) GetFromShort(k config.ShortKey) (v string, err error) {
	r.mg.RLock()
	defer r.mg.RUnlock()
	if item, ok := r.Data[k]; !ok {
		err = errors.ErrNotExist
	} else {
		v = item.url
	}
	return
}

func (r *MemStorageRepository) GetAll() Store {
	return r.Data
}

func (r *MemStorageRepository) RestoreAll(data Store) {
	r.Data = data
}
