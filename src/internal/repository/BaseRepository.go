package repository

import (
	"context"
	"database/sql"
	"fmt"
	sq "github.com/Masterminds/squirrel"
	"webinar2/src/internal/service/database"
)

const storageTable = "storage"

var storageColumns = []string{
	"key",
	"value",
}

type log interface {
	Info(args ...interface{})
}

type BaseRepository struct {
	db     database.QueryExecutor
	logger log
}

func NewBaseRepository(db database.QueryExecutor, loger log) *BaseRepository {
	return &BaseRepository{
		db:     db,
		logger: loger,
	}
}

func (br *BaseRepository) Get(key string) (string, error) {
	query, params, err := sq.Select(storageColumns...).
		From(storageTable).
		Where(sq.Eq{"key": key}).
		ToSql()
	if err != nil {
		return "", fmt.Errorf("query making: %w", err)
	}
	query = br.db.Rebind(query)
	br.logger.Info(fmt.Sprintf("query %s with params %#v", query, params))
	res, err := br.db.QueryContext(context.Background(), query, params...)
	if err != nil {
		return "", fmt.Errorf("result error: %w", err)
	}
	if !res.Next() {
		return "", sql.ErrNoRows
	}
	var name, value string
	br.logger.Info(fmt.Sprintf("key: %s has value: %s", name, value))
	err = res.Scan(&name, &value)
	switch {
	case err == sql.ErrNoRows:
		return "", err
	case err != nil:
		return "", fmt.Errorf("recieve value %w", err)
	}

	return value, nil
}

func (br *BaseRepository) Set(key, value string) error {
	query, params, err := sq.Insert(storageTable).
		Columns(storageColumns...).
		Values(key, value).
		Suffix("ON CONFLICT(key) DO UPDATE SET value = EXCLUDED.value").ToSql()
	if err != nil {
		return fmt.Errorf("query making: %w", err)
	}

	query = br.db.Rebind(query)
	br.logger.Info(fmt.Sprintf("query %s with params %#v", query, params))
	_, err = br.db.Exec(query, params...)
	if err != nil {
		return fmt.Errorf("insertion error: %w", err)
	}
	return nil
}
