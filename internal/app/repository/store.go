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

type Store map[domain.ShortKey]*storeItem

func itfToString(i interface{}) (s string) {
	s, _ = i.(string)
	return
}

func newStoreItem(ctx context.Context, attrs ...interface{}) *storeItem {
	att := make([]interface{}, 4)
	copy(att, attrs)
	if att[2] == nil {
		u, ok := ctx.Value(constant.ContextUserValueName).(string)
		if ok {
			att[2] = u
		}
	}

	return &storeItem{
		uuid:   itfToString(att[0]),
		url:    itfToString(att[1]),
		userID: itfToString(att[2]),
		isDeleted: func(i interface{}) (b bool) {
			b, _ = i.(bool)
			return
		}(att[3]),
	}
}
