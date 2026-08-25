package mysql

import (
	"context"
	"database/sql"
	"net/url"
	"strings"

	mysqldriver "github.com/go-sql-driver/mysql"

	"godruid/internal/connx"
)

// Connector creates one sql.Conn at a time. The backing sql.DB is forced to
// MaxOpen=1 and MaxIdle=0 so the driver cannot nest an uncontrolled pool.
type Connector struct {
	DSN string
}

func New(dsn string) *Connector { return &Connector{DSN: dsn} }

func (c *Connector) Kind() string { return "mysql" }

func (c *Connector) Connect(ctx context.Context) (connx.Connection, error) {
	db, err := sql.Open("mysql", c.DSN)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
	db.SetConnMaxLifetime(0)
	sc, err := db.Conn(ctx)
	if err != nil {
		_ = db.Close()
		return nil, err
	}
	return &Conn{db: db, conn: sc, remote: sanitizeDSN(c.DSN)}, nil
}

type Conn struct {
	db     *sql.DB
	conn   *sql.Conn
	remote string
}

func (c *Conn) Ping(ctx context.Context) error {
	return c.conn.PingContext(ctx)
}

func (c *Conn) Close() error {
	var err error
	if c.conn != nil {
		err = c.conn.Close()
	}
	if c.db != nil {
		if e := c.db.Close(); e != nil && err == nil {
			err = e
		}
	}
	return err
}

func (c *Conn) Metadata() connx.Metadata {
	return connx.Metadata{
		Kind:     "mysql",
		Remote:   c.remote,
		Isolated: true,
		Note:     "sql.DB MaxOpen=1 MaxIdle=0; do not share this DB",
	}
}

func sanitizeDSN(dsn string) string {
	if cfg, err := mysqldriver.ParseDSN(dsn); err == nil {
		host := cfg.Addr
		if cfg.DBName != "" {
			return host + "/" + cfg.DBName
		}
		return host
	}
	if u, err := url.Parse(dsn); err == nil {
		u.User = nil
		return u.Host
	}
	if i := strings.Index(dsn, "@"); i >= 0 {
		return dsn[i+1:]
	}
	return "mysql"
}

// ApplyIsolation documents and applies the single-connection constraints.
func ApplyIsolation(db *sql.DB) {
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(0)
}
