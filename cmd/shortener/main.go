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

	logger := logrus.New()
	logger.WithFields(logrus.Fields{
		"Server Address": conf.ServerAddress,
		"Base URL":       conf.BaseURL,
	}).Info("Start server")

	r := repository.NewRepository(conf.FileStoragePath)
	s := service.NewService(r, conf)

	if data, err := r.Restore(); err != nil {
		logger.WithError(err).Error("Can not restore data")
	} else {
		r.RestoreAll(data)
	}

	defer func() {
		if err := r.Save(r.GetAll()); err != nil {
			logger.WithError(err).Error("Can not save data")
		}
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
