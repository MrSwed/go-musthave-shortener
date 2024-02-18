package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/MrSwed/go-musthave-shortener/internal/app/config"
	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
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
	db *sqlx.DB
}

func NewDBStorageRepository(db *sqlx.DB) *DBStorageRepo {
	return &DBStorageRepo{
		db: db,
	}
}

func (r *DBStorageRepo) Ping() error {
	if r.db == nil {
		return fmt.Errorf("no DB connected")
	}
	return r.db.Ping()
}

func (r *DBStorageRepo) saveNew(item DBStorageItem) (err error) {
	_, err = r.db.Exec("insert into "+constant.DBTableName+" (short, url) values ($1, $2)",
		item.Short, item.URL)
	return
}

func (r *DBStorageRepo) NewShort(ctx context.Context, url string) (short string, err error) {
	for {
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		default:
			newShort := helper.NewRandShorter().RandStringBytes().String()
			if errS := r.saveNew(DBStorageItem{Short: newShort, URL: url}); errS == nil {
				short = newShort
				return
			} else if errP, ok := errS.(*pgconn.PgError); !ok || errP.Code != pgerrcode.UniqueViolation {
				err = errS
				return
			}
		}
	}
}

func (r *DBStorageRepo) GetFromShort(ctx context.Context, k string) (v string, err error) {
	if len([]byte(k)) != len(config.ShortKey{}) {
		err = myErr.ErrNotExist
		return
	}
	sqlStr := `SELECT uuid, short, url FROM ` + constant.DBTableName + ` WHERE short = $1`
	var item = DBStorageItem{}
	if err = r.db.GetContext(ctx, &item, sqlStr, k); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
		return
	}
	v = item.URL
	return
}

func (r *DBStorageRepo) GetFromURL(ctx context.Context, url string) (v string, err error) {
	var item = DBStorageItem{}
	sqlStr := `SELECT uuid, short, url FROM ` + constant.DBTableName + ` WHERE url = $1`
	if err = r.db.GetContext(ctx, &item, sqlStr, url); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return
	}
	v = item.Short
	return
}

func (r *DBStorageRepo) GetAll(ctx context.Context) (data Store, err error) {
	data = make(Store)
	sqlStr := `SELECT uuid, short, url FROM ` + constant.DBTableName
	var rows *sql.Rows
	if rows, err = r.db.QueryContext(ctx, sqlStr); err != nil {
		return
	}
	defer func() { err = rows.Close() }()
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
	err = rows.Err()
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

func (r *DBStorageRepo) NewShortBatch(ctx context.Context, input []domain.ShortBatchInputItem, prefix string) (out []domain.ShortBatchResultItem, err error) {
	var (
		tx *sqlx.Tx
	)
	tx, err = r.db.Beginx()
	if err != nil {
		return
	}
	defer func() {
		rErr := tx.Rollback()
		if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			err = errors.Join(err, rErr)
			out = nil
		}
	}()

	for _, i := range input {
		for {
			var newShort string
			if err = tx.GetContext(ctx, &newShort, "select short from "+constant.DBTableName+" where url = $1", i.OriginalURL); err == nil {
				out = append(out, domain.ShortBatchResultItem{
					CorrelationTD: i.CorrelationID,
					ShortURL:      prefix + newShort,
				})
				break
			}
			newShort = helper.NewRandShorter().RandStringBytes().String()
			var exist int
			if err = tx.GetContext(ctx, &exist, "select count(short) from "+constant.DBTableName+" where short = $1", newShort); err != nil {
				return
			}
			if exist > 0 {
				continue
			}
			if _, err = tx.ExecContext(ctx, "INSERT INTO "+constant.DBTableName+" (short, url) VALUES($1, $2)", newShort, i.OriginalURL); err == nil {
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
	err = errors.Join(err, tx.Commit())
	return
}
