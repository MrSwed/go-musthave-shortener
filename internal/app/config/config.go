package config

const (
	Scheme         = "http://"
	Address        = "localhost:8080"
	MakeShortRoute = "/"
	ShortLen       = 8
)

type ShortKey [ShortLen]byte
