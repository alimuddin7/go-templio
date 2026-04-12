package migrations

import (
	"embed"

	"github.com/uptrace/bun/migrate"
)

//go:embed *.sql
var sqlMigrations embed.FS

// Migrations is the bun/migrate structure representing the embedded filesystem.
var Migrations = migrate.NewMigrations()

func init() {
	if err := Migrations.DiscoverCaller(); err != nil {
		panic(err)
	}
	if err := Migrations.Discover(sqlMigrations); err != nil {
		panic(err)
	}
}
