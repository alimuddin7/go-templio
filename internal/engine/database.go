package engine

import (
	"context"
	"database/sql"
	"fmt"

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
		sqldb, err = sql.Open(sqliteshim.ShimName, cfg.DSN)
		if err != nil {
			return nil, fmt.Errorf("database: open sqlite: %w", err)
		}
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
