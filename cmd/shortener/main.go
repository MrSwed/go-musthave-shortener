package main

import (
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/handler"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"
	"log"
)

func main() {
	conf := config.NewConfig().Init()

	log.Printf(`Started with config:
  Server Address: %s
  Base Url: %s
`, conf.ServerAddress, conf.BaseURL)

	r := repository.NewRepository()
	s := service.NewService(r, conf)

	handler.NewHandler(s).RunServer(conf.ServerAddress)
}
