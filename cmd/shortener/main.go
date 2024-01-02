package main

import (
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/handler"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"
)

func main() {

	r := repository.NewRepository()
	s := service.NewService(r)

	handler.NewHandler(s).RunServer(config.Address)
}
