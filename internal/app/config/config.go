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

func NewConfig() *Config {
	c := &Config{ServerAddress, BaseURL, Scheme}
	return c
}

func (c *Config) WithFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Provide base address for short url")
	flag.Parse()
	return c
}
