package config

import (
	"flag"
	"os"
	"strings"
)

type ShortKey [ShortLen]byte

func (s ShortKey) String() string {
	return string(s[:])
}

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	Scheme          string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress:   serverAddress,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		Scheme:          scheme,
	}
}

func (c *Config) Init() *Config {
	return c.withFlags().withEnv().cleanSchemes()
}

func (c *Config) withEnv() *Config {
	if envAddress, ok := os.LookupEnv(envServerAddressName); ok && envAddress != "" {
		c.ServerAddress = envAddress
	}
	if envBaseURL, ok := os.LookupEnv(envBaseURLName); ok && envBaseURL != "" {
		c.BaseURL = envBaseURL
	}
	if envFileStoragePath, ok := os.LookupEnv(envFileStoragePathName); ok && envFileStoragePath != "" {
		c.FileStoragePath = envFileStoragePath
	}
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Provide base address for short url")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Provide storage file")
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
