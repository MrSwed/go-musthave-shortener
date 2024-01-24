package repository

import (
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/errors"
)

type Store map[config.ShortKey]string

type MemStorage interface {
	SaveShort(k config.ShortKey, v string) error
	GetFromShort(k config.ShortKey) (string, error)
	GetAll() Store
	RestoreAll(data Store)
}

type MemStorageRepository struct {
	Data Store
	mg   sync.RWMutex
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{
		Data: Store{},
	}
}

func (m *MemStorageRepository) SaveShort(k config.ShortKey, v string) (err error) {
	m.mg.Lock()
	defer m.mg.Unlock()
	m.Data[k] = v
	return
}

func (m *MemStorageRepository) GetFromShort(k config.ShortKey) (v string, err error) {
	var ok bool
	m.mg.RLock()
	defer m.mg.RUnlock()
	if v, ok = m.Data[k]; !ok {
		err = errors.ErrNotExist
	}
	return
}

func (m *MemStorageRepository) GetAll() Store {
	return m.Data
}

func (m *MemStorageRepository) RestoreAll(data Store) {
	m.Data = data
}
