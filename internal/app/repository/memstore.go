package repository

import (
	"context"
	"sync"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"

	"github.com/google/uuid"
)

type MemStorageRepository struct {
	Data Store
	mg   sync.RWMutex
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{
		Data: make(Store),
	}
}

func (r *MemStorageRepository) NewShort(ctx context.Context, url string) (short string, err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			newShort := helper.NewRandShorter().RandStringBytes()
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
}

func (r *MemStorageRepository) GetFromShort(ctx context.Context, k string) (v string, err error) {
	if len([]byte(k)) != len(config.ShortKey{}) {
		err = myErr.ErrNotExist
		return
	}
	sk := config.ShortKey([]byte(k))
	r.mg.RLock()
	defer r.mg.RUnlock()
	if item, ok := r.Data[sk]; !ok {
		err = myErr.ErrNotExist
	} else {
		v = item.url
	}
	return
}

func (r *MemStorageRepository) GetFromURL(ctx context.Context, url string) (v string, err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	for sk, item := range r.Data {
		if item.url == url {
			return sk.String(), nil
		}
	}

	return
}

func (r *MemStorageRepository) GetAll(ctx context.Context) (Store, error) {
	return r.Data, nil
}

func (r *MemStorageRepository) RestoreAll(data Store) error {
	r.Data = data
	return nil
}

func (r *MemStorageRepository) NewShortBatch(ctx context.Context, input []domain.ShortBatchInputItem, prefix string) (out []domain.ShortBatchResultItem, err error) {
	for _, i := range input {
		var short string
		if short, err = r.GetFromURL(ctx, i.OriginalURL); err != nil {
			return
		}
		if short == "" {
			if short, err = r.NewShort(ctx, i.OriginalURL); err != nil {
				return
			}
		}
		out = append(out, domain.ShortBatchResultItem{
			CorrelationTD: i.CorrelationID,
			ShortURL:      prefix + short,
		})
	}
	return
}
