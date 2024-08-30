package adaptor

import (
	"database/sql"
	"fmt"
	"github.com/Kiyo510/url-shorter/internal/config"
	"github.com/jmoiron/sqlx"
	_ "github.com/newrelic/go-agent/v3/integrations/nrpq"
	"sync"
)

func NewPostgresAdaptor() DBConn {
	var dba DBAdaptor
	var mtx sync.Mutex
	return func() (DBAdaptor, error) {

		mtx.Lock()
		defer mtx.Unlock()
		if dba != nil {
			return dba, nil
		}
		conf := config.PostgresConf
		dsn := fmt.Sprintf("user=%s password=%s host=%s dbname=%s sslmode=%s", conf.User, conf.Pass, conf.Host, conf.Name, "disable")
		db, err := sql.Open("nrpostgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("sqlx connect failed: %w", err)
		}
		dbx := sqlx.NewDb(db, "nrpostgres")
		dba = &sqlxConn{conn: dbx}
		return dba, nil
	}
}
