package repository

import "github.com/MrSwed/go-musthave-shortener/internal/app/config"

type storeItem struct {
	uuid string
	url  string
}

type Store map[config.ShortKey]storeItem
