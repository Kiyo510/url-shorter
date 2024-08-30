package adaptor

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DBConn func() (DBAdaptor, error)

// TODO: contextを受け取るようにする
type DBAdaptor interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
	GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error
}

type sqlxConn struct {
	conn *sqlx.DB
}

func (s *sqlxConn) Get(dest interface{}, query string, args ...interface{}) error {
	return s.conn.Get(dest, query, args...)
}

func (s *sqlxConn) Exec(query string, args ...interface{}) (sql.Result, error) {
	return s.conn.Exec(query, args...)
}

func (s *sqlxConn) GetContext(ctx context.Context, dest interface{}, query string, args ...interface{}) error {
	return s.conn.GetContext(ctx, dest, query, args...)
}
