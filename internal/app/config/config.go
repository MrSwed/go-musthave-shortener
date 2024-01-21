package config

import (
	"flag"
	"os"
	"strings"
)

type ShortKey [ShortLen]byte

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	Scheme          string
}

func NewConfig(init ...bool) *Config {
	c := &Config{
		ServerAddress:   serverAddress,
		BaseURL:         baseURL,
		FileStoragePath: fileStoragePath,
		Scheme:          scheme,
	}
	if len(init) > 0 && init[0] {
		return c.withFlags().withEnv().cleanSchemes()
	}
	return c
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
	flag.StringVar(&c.ServerAddress, "a", serverAddress, "Provide the address start server")
	flag.StringVar(&c.BaseURL, "b", baseURL, "Provide base address for short url")
	flag.StringVar(&c.FileStoragePath, "f", fileStoragePath, "Provide storage file")
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
