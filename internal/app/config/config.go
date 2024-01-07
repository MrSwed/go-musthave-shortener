package config

import (
	"flag"
	"strings"
)

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
	return c.cleanSchemes()
}

func (c *Config) cleanSchemes() *Config {
	c.ServerAddress = strings.TrimPrefix("http://", c.ServerAddress)
	c.ServerAddress = strings.TrimPrefix("https://", c.ServerAddress)
	c.BaseURL = strings.TrimPrefix("http://", c.BaseURL)
	c.BaseURL = strings.TrimPrefix("https://", c.BaseURL)
	return c
}
