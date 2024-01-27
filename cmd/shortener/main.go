package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/handler"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/sirupsen/logrus"
)

func main() {
	conf := config.NewConfig().Init()

	logger := logrus.New()
	logger.WithFields(logrus.Fields{
		"Server Address":  conf.ServerAddress,
		"Base URL":        conf.BaseURL,
		"FileStoragePath": conf.FileStoragePath,
	}).Info("Start server")

	r := repository.NewRepository(conf)
	s := service.NewService(r)
	h := handler.NewHandler(s, logger)

	if conf.FileStoragePath != "" {
		data, err := r.Restore()
		if err != nil {
			logger.WithError(err).Error("Storage restore")
		}
		if data != nil {
			r.RestoreAll(data)
			logger.Info("Storage restored")
		}
	}

	server := &http.Server{
		Addr:    conf.ServerAddress,
		Handler: h.Handler(),
	}

	serverCtx, serverStop := context.WithCancel(context.Background())

	exitSig := make(chan os.Signal, 1)
	signal.Notify(exitSig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		<-exitSig

		shutdownCtx, shutdownStopForce := context.WithTimeout(serverCtx, config.ServerShutdownTimeout*time.Second)
		defer shutdownStopForce()
		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				logger.Error("graceful shutdown timed out.. forcing exit.")
			}
		}()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.WithError(err).Error("shutdown")
		}
		serverStop()
	}()

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.WithError(err).Error("Start server")
		serverStop()
	}

	<-serverCtx.Done()

	if conf.FileStoragePath != "" {
		if err := r.FileStorage.Save(r.MemStorage.GetAll()); err != nil {
			logger.WithError(err).Error("Can not save data")
		} else {
			logger.Info("Storage saved")
		}
	}
	logger.Info("Server stopped")

}
