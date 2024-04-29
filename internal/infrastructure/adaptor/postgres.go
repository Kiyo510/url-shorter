package adaptor

import (
	"database/sql"
	"fmt"

	"github.com/Kiyo510/url-shorter/internal/config"
	"github.com/jmoiron/sqlx"
)

// TODO: contextを受け取るようにする
type DBAdapter interface {
	Get(dest interface{}, query string, args ...interface{}) error
	Exec(query string, args ...interface{}) (sql.Result, error)
}

type postgresAdapter struct {
	db *sqlx.DB
}

func (p *postgresAdapter) Exec(query string, args ...interface{}) (sql.Result, error) {
	return p.db.Exec(query, args...)
}

func (p *postgresAdapter) Get(dest interface{}, query string, args ...interface{}) error {
	return p.db.Get(dest, query, args...)
}

// TODO: contextを受け取るために、関数型を返すようにしても良いかも
func NewPostgresAdapter() (DBAdapter, error) {
	conf := config.PostgresConf
	dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s", conf.User, conf.Pass, conf.Host, conf.Name, "disable")
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("sqlx connect failed: %w", err)
	}

	return &postgresAdapter{db: db}, nil
}
