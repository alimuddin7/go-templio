package engine

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/pgdialect"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/driver/pgdriver"
	"github.com/uptrace/bun/driver/sqliteshim"
	"github.com/uptrace/bun/extra/bundebug"

	"github.com/alimuddin7/go-templio/internal/config"
)

// Database wraps a *bun.DB and exposes lifecycle helpers.
type Database struct {
	DB *bun.DB
}

// NewDatabase creates and validates a Bun connection based on the driver setting.
// Supported drivers: "sqlite", "postgres".
func NewDatabase(cfg config.DatabaseConfig, debug bool) (*Database, error) {
	var (
		sqldb *sql.DB
		bundb *bun.DB
	)

	switch cfg.Driver {
	case "sqlite":
		var err error
		// Enable foreign keys and set busy timeout
		dsn := cfg.DSN
		if !strings.Contains(dsn, "?") {
			dsn += "?_foreign_keys=1&_busy_timeout=5000"
		} else if !strings.Contains(dsn, "_foreign_keys") {
			dsn += "&_foreign_keys=1&_busy_timeout=5000"
		}

		sqldb, err = sql.Open(sqliteshim.ShimName, dsn)
		if err != nil {
			return nil, fmt.Errorf("database: open sqlite: %w", err)
		}
		// SQLite only supports one writer at a time.
		sqldb.SetMaxOpenConns(1)

		bundb = bun.NewDB(sqldb, sqlitedialect.New())

	case "postgres":
		connector := pgdriver.NewConnector(
			pgdriver.WithDSN(cfg.PostgresDSN()),
		)
		sqldb = sql.OpenDB(connector)
		bundb = bun.NewDB(sqldb, pgdialect.New())

	default:
		return nil, fmt.Errorf("database: unsupported driver %q", cfg.Driver)
	}

	if debug {
		bundb.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	// Validate connection.
	if err := bundb.PingContext(context.Background()); err != nil {
		return nil, fmt.Errorf("database: ping failed: %w", err)
	}

	return &Database{DB: bundb}, nil
}

// Close releases the underlying SQL connection pool.
func (d *Database) Close() error {
	return d.DB.Close()
}
