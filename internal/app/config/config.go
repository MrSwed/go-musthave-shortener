package config

import (
	"flag"
	"os"
	"strings"
)

const (
	Scheme        = "http://"
	ServerAddress = "localhost:8080"
	BaseURL       = "localhost:8080"
	ShortLen      = 8

	APIRoute     = "/api"
	ShortenRoute = "/shorten"
)

type ShortKey [ShortLen]byte

type Config struct {
	ServerAddress string
	BaseURL       string
	Scheme        string
}

func NewConfig(init ...bool) *Config {
	c := &Config{ServerAddress, BaseURL, Scheme}
	if len(init) > 0 && init[0] {
		return c.withFlags().withEnv().cleanSchemes()
	}
	return c
}

func (c *Config) withEnv() *Config {
	serverAddress, baseURL := os.Getenv("SERVER_ADDRESS"), os.Getenv("BASE_URL")
	if serverAddress != "" {
		c.ServerAddress = serverAddress
	}
	if baseURL != "" {
		c.BaseURL = baseURL
	}
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Provide base address for short url")
	flag.Parse()
	return c
}

func (c *Config) cleanSchemes() *Config {
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "http://")
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "https://")
	c.BaseURL = strings.TrimPrefix(c.BaseURL, "http://")
	c.BaseURL = strings.TrimPrefix(c.BaseURL, "https://")
	return c
}
