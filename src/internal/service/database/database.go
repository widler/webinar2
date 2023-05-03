package database

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // nolint: golint
)

const (
	dbDriver = "postgres"
)

type Logger interface {
	Info(args ...interface{})
}

type Database struct {
	db     *sqlx.DB
	Logger Logger
}

type QueryExecutor interface {
	Rebind(string) string
	sqlx.Execer
	sqlx.QueryerContext
}

func NewDatabase(dsn string, log Logger) (*Database, error) {
	db, err := sqlx.Open(dbDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("connection: %w", err)
	}

	return &Database{
		db:     db,
		Logger: log,
	}, nil
}

func (db Database) DB() *sqlx.DB {
	return db.db
}

func (db Database) Close() error {
	return db.db.Close()
}
