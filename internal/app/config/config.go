package config

import "flag"

const (
	Scheme        = "http://"
	ServerAddress = "localhost:8080"
	BaseURL       = "localhost:8080"
	ShortLen      = 8
)

type ShortKey [ShortLen]byte

type Config struct {
	ServerAddress string
	BaseURL       string
	Scheme        string
}

func (c *Config) InitConfig() {
	flag.StringVar(&c.ServerAddress, "a", ServerAddress, "Provide the address start server")
	flag.StringVar(&c.BaseURL, "b", BaseURL, "Provide base address for short url")
	c.Scheme = Scheme
	flag.Parse()
}
