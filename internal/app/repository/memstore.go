package repository

import (
	"fmt"
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/errors"
)

type storeItem struct {
	uuid string
	url  string
}

type Store map[config.ShortKey]storeItem

type MemStorage interface {
	SaveShort(k config.ShortKey, v string) error
	GetFromShort(k config.ShortKey) (string, error)
	GetAll() Store
	RestoreAll(Store)
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
	m.Data[k] = storeItem{
		// tmp use int instead uuid
		uuid: fmt.Sprint(len(m.Data) + 1),
		url:  v,
	}
	return
}

func (m *MemStorageRepository) GetFromShort(k config.ShortKey) (v string, err error) {
	m.mg.RLock()
	defer m.mg.RUnlock()
	if item, ok := m.Data[k]; !ok {
		err = errors.ErrNotExist
	} else {
		v = item.url
	}
	return
}

func (m *MemStorageRepository) GetAll() Store {
	return m.Data
}

func (m *MemStorageRepository) RestoreAll(data Store) {
	m.Data = data
}
