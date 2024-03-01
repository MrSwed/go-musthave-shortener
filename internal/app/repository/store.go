package repository

import (
	"context"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"

	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
)

type storeItem struct {
	uuid      string
	url       string
	userID    string
	isDeleted bool
}

type Store map[domain.ShortKey]storeItem

func newStoreItem(ctx context.Context, attrs ...string) storeItem {
	att := make([]string, 3)
	copy(att, attrs)
	if att[2] == "" {
		if u, ok := ctx.Value(constant.ContextUserValueName).(string); ok {
			att[2] = u
		}
	}

	return storeItem{
		uuid:   att[0],
		url:    att[1],
		userID: att[2],
	}
}
