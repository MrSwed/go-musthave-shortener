package repository

import (
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"

	"github.com/google/uuid"
)

type DataStorage interface {
	GetFromShort(k string) (string, error)
	NewShort(url string) (newURL string, err error)
	GetAll() (Store, error)
	RestoreAll(Store) error
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
				uuid: uuid.New().String(),
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

func (r *MemStorageRepository) GetAll() (Store, error) {
	return r.Data, nil
}

func (r *MemStorageRepository) RestoreAll(data Store) error {
	r.Data = data
	return nil
}
