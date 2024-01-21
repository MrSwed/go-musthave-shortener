package repository

import (
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/errors"
)

type MemStorage interface {
	SaveShort(k config.ShortKey, v string) error
	GetFromShort(k config.ShortKey) (string, error)
	GetAll() map[config.ShortKey]string
	RestoreAll(data map[config.ShortKey]string)
}

type MemStorageRepository struct {
	Data map[config.ShortKey]string
	mg   sync.RWMutex
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{
		Data: map[config.ShortKey]string{},
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

func (m *MemStorageRepository) GetAll() map[config.ShortKey]string {
	return m.Data
}

func (m *MemStorageRepository) RestoreAll(data map[config.ShortKey]string) {
	m.Data = data
}
