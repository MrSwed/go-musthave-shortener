package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/handler"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/sirupsen/logrus"
)

func main() {
	conf := config.NewConfig(true)

	r := repository.NewRepository()
	s := service.NewService(r, conf)
	logger := logrus.New()

	logger.WithFields(logrus.Fields{
		"Server Address": conf.ServerAddress,
		"Base Url":       conf.BaseURL,
	}).Info("Start server")

	defer func() {
		logger.Info("Server stopped")
	}()

	go func() {
		if err := handler.NewHandler(s, logger).RunServer(conf.ServerAddress); err != nil {
			logger.WithError(err).Error("Can not start server")
			os.Exit(1)
		}
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c
}
