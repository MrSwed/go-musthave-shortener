package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
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
	for {
		newShort := helper.NewRandShorter().RandStringBytes()
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
	row := r.db.QueryRow(context.Background(), sqlStr, k)
	var item = DBStorageItem{}
	if err = row.Scan(&item.UUID, &item.Short, &item.URL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = myErrs.ErrNotExist
		}
		return
	}
	v = item.URL
	return
}

func (r *DBStorageRepo) GetFromURL(url string) (v string, err error) {
	sqlStr := `SELECT uuid, short, url FROM ` + config.DBTableName + ` WHERE url = $1`
	row := r.db.QueryRow(context.Background(), sqlStr, url)

	var item = DBStorageItem{}
	if err = row.Scan(&item.UUID, &item.Short, &item.URL); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			err = nil
		}
		return
	}
	v = item.Short
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

func (r *DBStorageRepo) NewShortBatch(input []domain.ShortBatchInputItem, prefix string) (out []domain.ShortBatchResultItem, err error) {
	var (
		tx  pgx.Tx
		ctx = context.Background()
	)
	tx, err = r.db.Begin(ctx)
	if err != nil {
		return
	}
	defer func() {
		rErr := tx.Rollback(ctx)
		if rErr != nil && !errors.Is(rErr, pgx.ErrTxClosed) {
			err = errors.Join(err, rErr)
			out = nil
		}
	}()

	for _, i := range input {
		for {
			var newShort string
			row := tx.QueryRow(ctx, "select short from "+config.DBTableName+" where url = $1", i.OriginalURL)
			if err = row.Scan(&newShort); err == nil {
				out = append(out, domain.ShortBatchResultItem{
					CorrelationTD: i.CorrelationID,
					ShortURL:      prefix + newShort,
				})
				break
			}
			newShort = helper.NewRandShorter().RandStringBytes().String()
			row = tx.QueryRow(ctx, "select count(short) from "+config.DBTableName+" where short = $1", newShort)
			var exist int
			if err = row.Scan(&exist); err != nil {
				return
			}
			if exist > 0 {
				continue
			}
			if _, err = tx.Exec(ctx, "INSERT INTO "+config.DBTableName+" (short, url) VALUES($1, $2)", newShort, i.OriginalURL); err == nil {
				out = append(out, domain.ShortBatchResultItem{
					CorrelationTD: i.CorrelationID,
					ShortURL:      prefix + newShort,
				})
				break
			} else {
				return
			}
		}
	}
	err = errors.Join(err, tx.Commit(ctx))
	return
}
