package config

import (
	"flag"
	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
	"os"
	"strings"
)

type ShortKey [constant.ShortLen]byte

func (s ShortKey) String() string {
	return string(s[:])
}

type Config struct {
	ServerAddress   string
	BaseURL         string
	FileStoragePath string
	DatabaseDSN     string
	Scheme          string
}

func NewConfig() *Config {
	return &Config{
		ServerAddress:   constant.ServerAddress,
		BaseURL:         constant.BaseURL,
		FileStoragePath: constant.FileStoragePath,
		Scheme:          constant.Scheme,
	}
}

func (c *Config) Init() *Config {
	return c.withFlags().WithEnv().CleanParameters()
}

func (c *Config) WithEnv() *Config {
	if envAddress, ok := os.LookupEnv(constant.EnvServerAddressName); ok && envAddress != "" {
		c.ServerAddress = envAddress
	}
	if envBaseURL, ok := os.LookupEnv(constant.EnvBaseURLName); ok && envBaseURL != "" {
		c.BaseURL = envBaseURL
	}
	if envFileStoragePath, ok := os.LookupEnv(constant.EnvFileStoragePathName); ok && envFileStoragePath != "" {
		c.FileStoragePath = envFileStoragePath
	}
	if dbDSN, ok := os.LookupEnv(constant.EnvNameDBDSN); ok {
		c.DatabaseDSN = dbDSN
	}
	return c
}

func (c *Config) withFlags() *Config {
	flag.StringVar(&c.ServerAddress, "a", c.ServerAddress, "Provide the address start server")
	flag.StringVar(&c.BaseURL, "b", c.BaseURL, "Provide base address for short url")
	flag.StringVar(&c.FileStoragePath, "f", c.FileStoragePath, "Provide storage file")
	flag.StringVar(&c.DatabaseDSN, "d", c.DatabaseDSN, "Provide the database dsn connect string")
	flag.Parse()
	return c
}

func (c *Config) CleanParameters() *Config {
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "http://")
	c.ServerAddress = strings.TrimPrefix(c.ServerAddress, "https://")
	c.BaseURL = strings.TrimPrefix(c.BaseURL, "http://")
	c.BaseURL = strings.TrimPrefix(c.BaseURL, "https://")
	c.DatabaseDSN = strings.Trim(c.DatabaseDSN, "'")
	return c
}
