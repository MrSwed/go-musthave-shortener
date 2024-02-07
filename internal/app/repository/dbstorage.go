package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	myErrs "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/lib/pq"
)

type DBStorageItem struct {
	UUID  string `db:"uuid"`
	Short string `db:"short"`
	URL   string `db:"url"`
}

type DBStorage interface {
	Ping() error
}

type DBStorageRepo struct {
	db *pgxpool.Pool
}

func NewDBStorageRepository(db *pgxpool.Pool) *DBStorageRepo {
	return &DBStorageRepo{
		db: db,
	}
}

func (r *DBStorageRepo) Ping() error {
	if r.db == nil {
		return fmt.Errorf("no DB connected")
	}
	return r.db.Ping(context.Background())
}

func (r *DBStorageRepo) saveNew(item DBStorageItem) (err error) {
	_, err = r.db.Exec(context.Background(),
		"insert into "+config.DBTableName+" (short, url) values ($1, $2)",
		item.Short, item.URL)
	return
}

func (r *DBStorageRepo) NewShort(url string) (short string, err error) {
	for newShort := helper.NewRandShorter().RandStringBytes(); ; {
		if errS := r.saveNew(DBStorageItem{Short: newShort.String(), URL: url}); errS == nil {
			short = newShort.String()
			break
		} else if errP, ok := errS.(*pgconn.PgError); !ok || errP.Code != pgerrcode.UniqueViolation {
			err = errS
			break
		}
	}
	return
}

func (r *DBStorageRepo) GetFromShort(k string) (v string, err error) {
	if len([]byte(k)) != len(config.ShortKey{}) {
		err = myErrs.ErrNotExist
		return
	}

	sqlStr := `SELECT uuid, short, url FROM ` + config.DBTableName + ` WHERE short = $1`
	row := r.db.QueryRow(context.Background(), sqlStr)
	var item = DBStorageItem{}
	if err = row.Scan(&item); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = myErrs.ErrNotExist
		}
		return
	}
	v = item.URL
	return
}

func (r *DBStorageRepo) GetAll() (data Store, err error) {
	data = make(Store)
	sqlStr := `SELECT uuid, short, url FROM ` + config.DBTableName
	var rows pgx.Rows
	if rows, err = r.db.Query(context.Background(), sqlStr); err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		var item = DBStorageItem{}
		if err = rows.Scan(&item.UUID, &item.Short, &item.URL); err != nil {
			return
		}
		data[config.ShortKey([]byte(item.Short))] = storeItem{
			uuid: item.UUID,
			url:  item.URL,
		}
	}
	return
}

func (r *DBStorageRepo) RestoreAll(data Store) (err error) {
	for short, item := range data {
		if err = r.saveNew(DBStorageItem{Short: short.String(), URL: item.url, UUID: item.uuid}); err != nil {
			return err
		}
	}
	return
}

/**/
