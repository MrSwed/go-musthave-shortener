package main

import (
	"context"
	"errors"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/MrSwed/go-musthave-shortener/internal/app/closer"
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/handler"
	myMigrate "github.com/MrSwed/go-musthave-shortener/internal/app/migrate"
	"github.com/MrSwed/go-musthave-shortener/internal/app/repository"
	"github.com/MrSwed/go-musthave-shortener/internal/app/service"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	runServer(ctx)
}

func runServer(ctx context.Context) {
	conf := config.NewConfig().Init()
	var err error
	logger := logrus.New()
	logger.WithFields(logrus.Fields{"config": conf}).Info("Start server")

	var (
		db      *sqlx.DB
		isNewDB = true
		c       = &closer.Closer{}
	)
	if len(conf.DatabaseDSN) > 0 {
		if db, err = sqlx.Open("pgx", conf.DatabaseDSN); err != nil {
			logger.WithError(err).Fatal("cannot connect db")
		}
		logger.Info("DB connected")
		versions, errM := myMigrate.Migrate(db.DB)
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

	r := repository.NewRepository(repository.Config{StorageFile: conf.FileStoragePath, DB: db})
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

	c.Add("WEB", server.Shutdown)
	if conf.FileStoragePath != "" {
		c.Add("FileStorage", func(ctx context.Context) error {
			if store, err := s.GetAll(); err != nil {
				logger.WithError(err).Error("Can get data for save to disk")
				return err
			} else if err := r.FileStorage.Save(store); err != nil {
				logger.WithError(err).Error("Can not save data")
				return err
			} else {
				logger.Info("Storage saved")
			}
			return nil
		})
	}
	if db != nil {
		c.Add("DB", func(ctx context.Context) (err error) {
			if err = db.Close(); err != nil {
				logger.WithError(err).Error("DB close")
			} else {
				logger.Info("Db Closed")
			}
			return
		})
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.WithError(err).Error("Start server")
		}
	}()
	logger.Info("Server started")
	<-ctx.Done()

	logger.Info("Shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), config.ServerShutdownTimeout*time.Second)
	defer cancel()

	if err = c.Close(shutdownCtx); err != nil {
		logger.Error(err)
	}

	logger.Info("Server stopped")
}
