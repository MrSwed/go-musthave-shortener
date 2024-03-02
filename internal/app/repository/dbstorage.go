package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"

	"github.com/MrSwed/go-musthave-shortener/internal/app/constant"
	"github.com/MrSwed/go-musthave-shortener/internal/app/domain"
	myErr "github.com/MrSwed/go-musthave-shortener/internal/app/errors"
	"github.com/MrSwed/go-musthave-shortener/internal/app/helper"

	sqrl "github.com/Masterminds/squirrel"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

var (
	sq = sqrl.StatementBuilder.PlaceholderFormat(sqrl.Dollar)
)

type DBStorageItem struct {
	UUID      string `db:"uuid"`
	Short     string `db:"short"`
	URL       string `db:"url"`
	UserID    string `db:"user_id,omitempty"`
	IsDeleted bool   `db:"is_deleted"`
}

func newDBStorageItem(ctx context.Context, attrs ...string) *DBStorageItem {
	att := make([]string, 3)
	copy(att, attrs)
	if att[2] == "" {
		if u, ok := ctx.Value(constant.ContextUserValueName).(string); ok {
			att[2] = u
		}
	}

	return &DBStorageItem{
		Short:  att[0],
		URL:    att[1],
		UserID: att[2],
	}
}

type DBStorageRepo struct {
	db *sqlx.DB
}

func NewDBStorageRepository(db *sqlx.DB) *DBStorageRepo {
	return &DBStorageRepo{
		db: db,
	}
}

func (r *DBStorageRepo) Ping(ctx context.Context) error {
	if r.db == nil {
		return fmt.Errorf("no DB connected")
	}
	return r.db.PingContext(ctx)
}

func (r *DBStorageRepo) saveNew(ctx context.Context, item *DBStorageItem) (err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Insert(constant.DBTableName).
		Columns("short", "url", "user_id").
		Values(item.Short, item.URL, item.UserID).ToSql(); err != nil {
		return
	}
	_, err = r.db.ExecContext(ctx, sqlStr, args...)
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
			if errS := r.saveNew(ctx, newDBStorageItem(ctx, newShort, url)); errS == nil {
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
	if !domain.IsValidShortKey(k) {
		err = myErr.ErrNotExist
		return
	}
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.
		Select("uuid", "short", "url", "user_id", "is_deleted").
		From(constant.DBTableName).
		Where(sqrl.Eq{"short": k}).
		ToSql(); err != nil {
		return
	}
	var item = DBStorageItem{}
	if err = r.db.GetContext(ctx, &item, sqlStr, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
		return
	}
	v = item.URL
	if item.IsDeleted {
		err = myErr.ErrIsDeleted
		return
	}
	return
}

func (r *DBStorageRepo) GetFromURL(ctx context.Context, url string) (v string, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.
		Select("uuid", "short", "url", "user_id", "is_deleted").
		From(constant.DBTableName).
		Where(sqrl.Eq{"url": url}).
		ToSql(); err != nil {
		return
	}
	var item = DBStorageItem{}
	if err = r.db.GetContext(ctx, &item, sqlStr, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = nil
		}
		return
	}
	v = item.Short
	if item.IsDeleted {
		err = myErr.ErrIsDeleted
	}
	return
}

func (r *DBStorageRepo) GetAll(ctx context.Context) (data Store, err error) {
	var (
		sqlStr string
	)
	if sqlStr, _, err = sq.
		Select("uuid", "short", "url", "user_id").
		From(constant.DBTableName).
		ToSql(); err != nil {
		return
	}

	data = make(Store)
	var rows *sql.Rows
	if rows, err = r.db.QueryContext(ctx, sqlStr); err != nil {
		return
	}
	defer func() { err = rows.Close() }()
	for rows.Next() {
		var item = DBStorageItem{}
		if err = rows.Scan(&item.UUID, &item.Short, &item.URL, &item.UserID); err != nil {
			return
		}
		sk, er := domain.NewShortKey(item.Short)
		if er != nil {
			continue
		}
		data[sk] = newStoreItem(ctx,
			uuid.New().String(),
			item.URL,
			item.UserID,
		)
	}
	err = rows.Err()
	return
}

func (r *DBStorageRepo) RestoreAll(data Store) (err error) {
	var tx *sqlx.Tx

	tx, err = r.db.Beginx()
	if err != nil {
		return
	}
	defer func() {
		rErr := tx.Rollback()
		if rErr != nil && !errors.Is(rErr, sql.ErrTxDone) {
			err = errors.Join(err, rErr)
		}
	}()

	for short, item := range data {
		var (
			sqlStr string
			args   []interface{}
		)

		if sqlStr, args, err = sq.
			Insert(constant.DBTableName).
			Columns("uuid", "short", "url", "user_id").
			Values(item.uuid, short, item.url, item.userID).
			ToSql(); err != nil {
			return
		}
		if _, err = tx.Exec(sqlStr, args...); err != nil {
			return
		}
		if item.userID != "" {
			if sqlStr, args, err = sq.
				Insert(constant.DBUsersTableName).
				Columns("id").
				Values(item.userID).
				Suffix("ON CONFLICT (id) DO nothing").
				ToSql(); err != nil {
				return
			}
			if _, err = tx.Exec(sqlStr, args...); err != nil {
				return
			}
		}
	}

	err = errors.Join(err, tx.Commit())
	return
}

func (r *DBStorageRepo) NewShortBatch(ctx context.Context, input []domain.ShortBatchInputItem, prefix string) (out []domain.ShortBatchResultItem, err error) {
	var (
		tx     *sqlx.Tx
		userID string
	)
	if u, ok := ctx.Value(constant.ContextUserValueName).(string); ok {
		userID = u
	}
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
			select {
			case <-ctx.Done():
				err = ctx.Err()
				return
			default:
			}
			var (
				sqlStr   string
				args     []interface{}
				newShort string
			)
			if sqlStr, args, err = sq.
				Select("short").
				From(constant.DBTableName).
				Where(sqrl.Eq{"url": i.OriginalURL}).
				ToSql(); err != nil {
				return
			}

			if err = tx.GetContext(ctx, &newShort, sqlStr, args...); err == nil {
				out = append(out, domain.ShortBatchResultItem{
					CorrelationTD: i.CorrelationID,
					ShortURL:      prefix + newShort,
				})
				break
			}
			newShort = helper.NewRandShorter().RandStringBytes().String()
			var exist int
			if sqlStr, args, err = sq.
				Select("count(short)").
				From(constant.DBTableName).
				Where(sqrl.Eq{"short": newShort}).
				ToSql(); err != nil {
				return
			}

			if err = tx.GetContext(ctx, &exist, sqlStr, args...); err != nil {
				return
			}
			if exist > 0 {
				continue
			}
			if sqlStr, args, err = sq.
				Insert(constant.DBTableName).
				Columns("short", "url", "user_id").
				Values(newShort, i.OriginalURL, userID).
				ToSql(); err != nil {
				return
			}
			if _, err = tx.ExecContext(ctx, sqlStr, args...); err == nil {
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

func (r *DBStorageRepo) GetUser(ctx context.Context, id string) (user domain.UserInfo, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if _, err = uuid.Parse(id); err != nil {
		err = myErr.ErrNotExist
		return
	}
	if sqlStr, args, err = sq.
		Select("id", "created_at").
		From(constant.DBUsersTableName).
		Where(sqrl.Eq{"id": id}).
		ToSql(); err != nil {
		return
	}

	if err = r.db.GetContext(ctx, &user, sqlStr, args...); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			err = myErr.ErrNotExist
		}
	}

	return
}

func (r *DBStorageRepo) NewUser(ctx context.Context) (id string, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.Insert(constant.DBUsersTableName).
		Columns("id").
		Values(sqrl.Expr("DEFAULT")).
		Suffix(`RETURNING "id"`).
		ToSql(); err != nil {
		return
	}
	err = r.db.GetContext(ctx, &id, sqlStr, args...)
	return
}

func (r *DBStorageRepo) GetAllByUser(ctx context.Context, userID, prefix string) (data []domain.StorageItem, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.
		Select("'"+prefix+"' || short as short", "url").
		From(constant.DBTableName).
		Where("user_id = $1 and is_deleted is not true", userID).
		ToSql(); err != nil {
		return
	}
	if err = r.db.SelectContext(ctx, &data, sqlStr, args...); err != nil {
		return
	}
	return
}

func (r *DBStorageRepo) SetDeleted(ctx context.Context, userID string, delete bool, shorts ...string) (n int64, err error) {
	var (
		sqlStr string
		args   []interface{}
	)
	if sqlStr, args, err = sq.
		Update(constant.DBTableName).
		Set("is_deleted", delete).
		Where(sqrl.Eq{"user_id": userID, "short": shorts}).
		ToSql(); err != nil {
		return
	}
	var result sql.Result
	if result, err = r.db.ExecContext(ctx, sqlStr, args...); err != nil {
		return
	}

	return result.RowsAffected()
}
