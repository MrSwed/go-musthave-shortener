package repository

import (
	"context"
	"sync"
	"time"

	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"

	"github.com/google/uuid"
)

type userData struct {
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type users map[string]userData

type MemStorageRepository struct {
	Data  Store
	Users users
	mg    sync.RWMutex
}

func NewMemRepository() *MemStorageRepository {
	return &MemStorageRepository{
		Data:  make(Store),
		Users: make(users),
	}
}

func (r *MemStorageRepository) Ping(ctx context.Context) (err error) {
	return
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
				r.Data[newShort] = newStoreItem(ctx,
					uuid.New().String(),
					url,
				)
				short = newShort.String()
				return
			}
		}
	}
}

func (r *MemStorageRepository) GetFromShort(ctx context.Context, k string) (v string, err error) {
	if len([]byte(k)) != len(domain.ShortKey{}) {
		err = myErr.ErrNotExist
		return
	}
	sk := domain.ShortKey([]byte(k))
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
	for _, v := range data {
		r.Users[v.userID] = userData{CreatedAt: time.Now()}
	}
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

func (r *MemStorageRepository) GetUser(ctx context.Context, id string) (user domain.UserInfo, err error) {
	if u, ok := r.Users[id]; ok {
		user = domain.UserInfo{
			ID:        id,
			CreatedAt: u.CreatedAt,
		}
	} else {
		err = myErr.ErrNotExist
	}
	return
}

func (r *MemStorageRepository) NewUser(ctx context.Context) (id string, err error) {
	id = uuid.NewString()
	r.Users[id] = userData{CreatedAt: time.Now()}
	return
}

func (r *MemStorageRepository) GetAllByUser(ctx context.Context, userID, prefix string) (data []domain.StorageItem, err error) {
	r.mg.Lock()
	defer r.mg.Unlock()
	for sk, item := range r.Data {
		if item.userID == userID {
			data = append(data, domain.StorageItem{
				ShortURL:    prefix + sk.String(),
				OriginalURL: item.url,
			})
		}
	}
	return
}
