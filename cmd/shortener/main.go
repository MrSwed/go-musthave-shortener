package main

import (
	"context"
	"errors"
	"github.com/golang-migrate/migrate/v4"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/handler"
	myMigrate "github.com/MrSwed/go-musthave-shortener/internal/app/migrate"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	conf := config.NewConfig().Init()
	var err error
	logger := logrus.New()
	logger.WithFields(logrus.Fields{
		"Server Address":  conf.ServerAddress,
		"Base URL":        conf.BaseURL,
		"FileStoragePath": conf.FileStoragePath,
	}).Info("Start server")

	var (
		db      *pgxpool.Pool
		isNewDB = true
	)
	if len(conf.DatabaseDSN) > 0 {
		if db, err = connectPostgres(conf.DatabaseDSN); err != nil {
			logger.WithError(err).Fatal("cannot connect db")
		}
		logger.Info("DB connected")
		versions, errM := myMigrate.Migrate(stdlib.OpenDBFromPool(db))
		switch {
		case errors.Is(errM, migrate.ErrNoChange):
			logger.Info("DB migrate: ", errM, versions)
		case errM == nil:
			logger.Info("DB migrate: new applied ", versions)
		default:
			logger.WithError(err).Fatal("DB migrate: ", versions)
		}
		isNewDB = versions[0] == 0
	}

	r := repository.NewRepositories(repository.Config{StorageFile: conf.FileStoragePath, DB: db})
	s := service.NewService(r, conf)
	h := handler.NewHandler(s, logger)

	if conf.FileStoragePath != "" && isNewDB {
		data, err := r.Restore()
		if err != nil {
			logger.WithError(err).Error("Storage restore")
		}
		if data != nil {
			if err = s.RestoreAll(data); err != nil {
				logger.Error(err)
			} else {
				logger.Info("Storage restored")
			}
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
		if store, err := s.GetAll(); err != nil {
			logger.WithError(err).Error("Can get data for save to disk")
		} else if err := r.FileStorage.Save(store); err != nil {
			logger.WithError(err).Error("Can not save data")
		} else {
			logger.Info("Storage saved")
		}
	}
	logger.Info("Server stopped")

}

func connectPostgres(sbDSN string) (db *pgxpool.Pool, err error) {
	var poolConfig *pgxpool.Config
	poolConfig, err = pgxpool.ParseConfig(sbDSN)
	if err != nil {
		return
	}

	db, err = pgxpool.NewWithConfig(context.Background(), poolConfig)
	return

}
