package domain

import (
	"errors"
	"time"

	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
)

type CreateURL struct {
	URL string `json:"url"`
}

type ResultURL struct {
	Result string `json:"result"`
}

type ShortBatchInputItem struct {
	CorrelationID string `json:"correlation_id" validate:"required"`
	OriginalURL   string `json:"original_url" validate:"required"`
}

type ShortBatchInput struct {
	List []ShortBatchInputItem `validate:"required,gt=0,dive"`
}

type ShortBatchResultItem struct {
	CorrelationTD string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

type StorageItem struct {
	ShortURL    string `json:"short_url" db:"short"`
	OriginalURL string `json:"original_url" db:"url"`
}

type UserInfo struct {
	ID        string    `json:"id" db:"id"`
	CreatedAt time.Time `json:"createdAt" db:"created_at"`
}

type ShortKey [constant.ShortLen]byte

func (s ShortKey) String() string {
	return string(s[:])
}

func NewShortKey(k string) (sk ShortKey, err error) {
	if !IsValidShortKey(k) {
		err = errors.New("not valid")
		return
	}
	sk = ShortKey([]byte(k))
	return
}

func IsValidShortKey(k string) bool {
	return len([]byte(k)) == len(ShortKey{})
}
