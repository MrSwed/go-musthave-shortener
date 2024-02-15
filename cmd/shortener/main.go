package main

import (
	"context"
	"errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
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
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: "2006-01-02 15:04:05",
		FullTimestamp:   true,
	})

	logrus.WithFields(logrus.Fields{"config": conf}).Info("Start server")

	var (
		db      *sqlx.DB
		isNewDB = true
		c       = &closer.Closer{}
	)
	if len(conf.DatabaseDSN) > 0 {
		if db, err = sqlx.Open("pgx", conf.DatabaseDSN); err != nil {
			logrus.WithError(err).Fatal("cannot connect db")
		}
		logrus.Info("DB connected")
		versions, errM := myMigrate.Migrate(db.DB)
		switch {
		case errors.Is(errM, migrate.ErrNoChange):
			logrus.Info("DB migrate: ", errM, versions)
		case errM == nil:
			logrus.Info("DB migrate: new applied ", versions)
		default:
			logrus.WithError(err).Fatal("DB migrate: ", versions)
		}
		isNewDB = versions[0] == 0
	}

	r := repository.NewRepository(repository.Config{StorageFile: conf.FileStoragePath, DB: db})
	s := service.NewService(r, conf)
	h := handler.NewHandler(s)

	if conf.FileStoragePath != "" && isNewDB {
		data, err := r.Restore()
		if err != nil {
			logrus.WithError(err).Error("Storage restore")
		}
		if data != nil {
			if err = s.RestoreAll(data); err != nil {
				logrus.Error(err)
			} else {
				logrus.Info("Storage restored")
			}
		}
	}

	server := &http.Server{
		Addr:    conf.ServerAddress,
		Handler: h.Handler(),
	}
	lockDBCLose := make(chan struct{})
	c.Add("WEB", server.Shutdown)
	if conf.FileStoragePath != "" {
		c.Add("FileStorage", func(ctx context.Context) error {
			defer func() {
				if db != nil {
					close(lockDBCLose)
				}
			}()
			if store, err := s.GetAll(ctx); err != nil {
				logrus.WithError(err).Error("Can get data for save to disk")
				return err
			} else if err := r.FileStorage.Save(store); err != nil {
				logrus.WithError(err).Error("Can not save data")
				return err
			} else {
				logrus.Info("Storage saved")
			}
			return nil
		})
	} else {
		close(lockDBCLose)
	}
	if db != nil {
		c.Add("DB", func(ctx context.Context) (err error) {
			<-lockDBCLose
			if err = db.Close(); err != nil {
				logrus.WithError(err).Error("DB close")
			} else {
				logrus.Info("Db Closed")
			}
			return
		})
	}

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logrus.WithError(err).Error("Start server")
		}
	}()
	logrus.Info("Server started")
	<-ctx.Done()

	logrus.Info("Shutting down server gracefully")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), constant.ServerShutdownTimeout*time.Second)
	defer cancel()

	if err = c.Close(shutdownCtx); err != nil {
		logrus.Error(err, ". timeout: ", constant.ServerShutdownTimeout)
	}

	logrus.Info("Server stopped")
}
