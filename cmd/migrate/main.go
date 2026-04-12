package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/alimuddin7/go-templio/internal/config"
	"github.com/alimuddin7/go-templio/internal/database/migrations"
	"github.com/alimuddin7/go-templio/internal/engine"

	"github.com/uptrace/bun/migrate"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	debug := cfg.App.Env != "production"
	db, err := engine.NewDatabase(cfg.Database, debug)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	migrator := migrate.NewMigrator(db.DB, migrations.Migrations)
	ctx := context.Background()

	cmd := os.Args[1]
	switch cmd {
	case "init":
		if err := migrator.Init(ctx); err != nil {
			log.Fatalf("init migrations failed: %v", err)
		}
		fmt.Println("✅ Migrations tracking initialized.")

	case "up":
		group, err := migrator.Migrate(ctx)
		if err != nil {
			log.Fatalf("migration failed: %v", err)
		}
		if group.IsZero() {
			fmt.Println("✅ Database is already up to date.")
		} else {
			fmt.Printf("✅ Migrated to %s\n", group)
		}

	case "down":
		group, err := migrator.Rollback(ctx)
		if err != nil {
			log.Fatalf("rollback failed: %v", err)
		}
		if group.IsZero() {
			fmt.Println("✅ No migrations to rollback.")
		} else {
			fmt.Printf("✅ Rolled back %s\n", group)
		}

	case "status":
		ms, err := migrator.MigrationsWithStatus(ctx)
		if err != nil {
			log.Fatalf("could not compute status: %v", err)
		}
		fmt.Printf("migrations: %s\n", ms)
		fmt.Printf("unapplied migrations: %s\n", ms.Unapplied())
		fmt.Printf("last migration group: %s\n", ms.LastGroup())

	case "create":
		if len(os.Args) < 3 {
			log.Fatal("missing migration name")
		}
		name := strings.Join(os.Args[2:], "_")
		files, err := migrator.CreateSQLMigrations(ctx, name)
		if err != nil {
			log.Fatalf("create failed: %v", err)
		}
		for _, f := range files {
			fmt.Printf("created migration %s (%s)\n", f.Name, f.Path)
		}

	default:
		fmt.Printf("unsupported command: %q\n", cmd)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println(`Usage: go run ./cmd/migrate [command] [args...]
Commands:
  init          Creates the bun_migrations table
  up            Runs all pending migrations
  down          Rolls back the last migration group
  status        Prints migration status
  create [name] Creates a new pair of UP/DOWN SQL migration files`)
}
