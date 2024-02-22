package helper

import (
	"crypto/rand"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	"math/big"
	mrand "math/rand"
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

func (r *RandShorter) RandStringBytes() domain.ShortKey {
	b := domain.ShortKey{}
	for i := range b {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(r.src)-1)))
		if err != nil {
			num = big.NewInt(int64(mrand.Intn(len(r.src) - 1)))
		}
		b[i] = r.src[num.Int64()]
	}
	return b
}
