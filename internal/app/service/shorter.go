package service

import (
	"context"
	"fmt"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"

	"github.com/go-playground/validator/v10"
)

type Shorter interface {
	NewShort(ctx context.Context, url string) (string, error)
	GetFromShort(ctx context.Context, k string) (string, error)
	CheckDB() error
	GetAll(ctx context.Context) (repository.Store, error)
	RestoreAll(repository.Store) error
	NewShortBatch(context.Context, []domain.ShortBatchInputItem) ([]domain.ShortBatchResultItem, error)
}

type ShorterService struct {
	r repository.Repository
	c *config.Config
}

func NewShorterService(r repository.Repository, c *config.Config) ShorterService {
	return ShorterService{r: r, c: c}
}

func (s ShorterService) fulNewShort(short string) string {
	return s.c.Scheme + s.c.BaseURL + "/" + short

}
func (s ShorterService) NewShort(ctx context.Context, url string) (newURL string, err error) {
	var newShort string
	if newShort, err = s.r.GetFromURL(ctx, url); err != nil {
		return
	}
	if newShort != "" {
		err = myErr.ErrAlreadyExist
	} else if newShort, err = s.r.NewShort(ctx, url); err != nil {
		return
	}

	newURL = s.fulNewShort(newShort)
	return
}

func (s ShorterService) GetFromShort(ctx context.Context, k string) (v string, err error) {
	v, err = s.r.GetFromShort(ctx, k)
	return
}

func (s ShorterService) CheckDB() (err error) {
	if rs, ok := s.r.(repository.Storage); ok {
		if rsDB, ok := rs.DataStorage.(repository.DBStorage); ok {
			err = rsDB.Ping()
			return
		}
	}
	return fmt.Errorf("no DB connected")
}

func (s ShorterService) GetAll(ctx context.Context) (repository.Store, error) {
	return s.r.GetAll(ctx)
}

func (s ShorterService) RestoreAll(data repository.Store) error {
	return s.r.RestoreAll(data)
}

func (s ShorterService) NewShortBatch(ctx context.Context, input []domain.ShortBatchInputItem) (out []domain.ShortBatchResultItem, err error) {
	validate := validator.New()
	if err = validate.Struct(domain.ShortBatchInput{List: input}); err != nil {
		return
	}

	return s.r.NewShortBatch(ctx, input, s.c.Scheme+s.c.BaseURL+"/")
}
