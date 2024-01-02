package helper

import (
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"math/rand"
)

type RandShorter struct {
	src []byte
}

func NewRandShorter(src ...byte) *RandShorter {
	if len(src) == 0 {
		src = []byte("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	}
	return &RandShorter{
		src: src,
	}
}

func (r *RandShorter) RandStringBytes() config.ShortKey {
	b := config.ShortKey{}
	for i := range b {
		b[i] = r.src[rand.Intn(len(r.src))]
	}
	return b
}
