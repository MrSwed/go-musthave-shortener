package repository

import (
	"fmt"
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"
)

type MemStorage interface {
	GetFromShort(k string) (string, error)
	NewShort(url string) (newURL string, err error)
	GetAll() Store
	RestoreAll(Store)
}

type MemStorageRepository struct {
	Data Store
	mg   sync.RWMutex
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{
		Data: make(Store),
	}
}

func (r *MemStorageRepository) NewShort(url string) (short string, err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	for newShort := helper.NewRandShorter().RandStringBytes(); ; {
		if _, exist := r.Data[newShort]; !exist {
			r.Data[newShort] = storeItem{
				// tmp use int instead uuid
				uuid: fmt.Sprint(len(r.Data) + 1),
				url:  url,
			}
			short = newShort.String()
			return
		}
	}
}

func (r *MemStorageRepository) GetFromShort(k string) (v string, err error) {
	if len([]byte(k)) != len(config.ShortKey{}) {
		err = errors.ErrNotExist
		return
	}
	sk := config.ShortKey([]byte(k))
	r.mg.RLock()
	defer r.mg.RUnlock()
	if item, ok := r.Data[sk]; !ok {
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
