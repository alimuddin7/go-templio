// Command server is the go-templio CMS application entrypoint.
// It wires together all layers and starts the Fiber HTTP server.
package main

import (
	"context"
	"flag"
	"log"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/crypto/bcrypt"
	bunmigrate "github.com/uptrace/bun/migrate"

	"github.com/alimuddin7/go-templio/internal/database/migrations"
	"github.com/alimuddin7/go-templio/internal/auth"
	"github.com/alimuddin7/go-templio/internal/config"
	"github.com/alimuddin7/go-templio/internal/domain/user"
	"github.com/alimuddin7/go-templio/internal/engine"
	"github.com/alimuddin7/go-templio/internal/navigation"
	"github.com/alimuddin7/go-templio/internal/transport/http/middleware"

	"github.com/alimuddin7/go-templio/internal/repository/user"
	usersvc "github.com/alimuddin7/go-templio/internal/service/user"
	userhandler "github.com/alimuddin7/go-templio/internal/transport/http/handler/user"

	authhandler "github.com/alimuddin7/go-templio/internal/transport/http/handler/auth"
	dashboardhandler "github.com/alimuddin7/go-templio/internal/transport/http/handler/dashboard"
	// [GEN-IMPORT]
)

func main() {
	migrate := flag.Bool("migrate", false, "run database migrations then exit")
	flag.Parse()

	// ── Config ────────────────────────────────────────────────────────────────
	cfg, err := config.Load(".env")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	// ── Database ──────────────────────────────────────────────────────────────
	debug := cfg.App.Env != "production"
	db, err := engine.NewDatabase(cfg.Database, debug)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// ── Navigation registry ───────────────────────────────────────────────────
	nav := navigation.NewRegistry()
	if err := nav.LoadYAML("navigation.yaml"); err != nil {
		log.Printf("navigation: failed to load YAML (using defaults): %v", err)
	}

	// ── Repositories ──────────────────────────────────────────────────────────
	userRepo := userrepo.New(db.DB)
	// [GEN-REPO]

	// ── Migrations ────────────────────────────────────────────────────────────
	ctx := context.Background()
	
	if *migrate {
		migrator := bunmigrate.NewMigrator(db.DB, migrations.Migrations)
		if err := migrator.Init(ctx); err != nil {
			log.Fatalf("migrate init: %v", err)
		}

		group, err := migrator.Migrate(ctx)
		if err != nil {
			log.Fatalf("migrate up: %v", err)
		}
		
		if group.IsZero() {
			log.Println("✅ No new migrations to run (database is up to date).")
		} else {
			log.Printf("✅ Migrated to %s\n", group)
		}
		return
	}


	// ── Seed default admin (idempotent) ───────────────────────────────────────
	seedAdmin(ctx, userRepo)

	// ── Services ──────────────────────────────────────────────────────────────
	userService := usersvc.New(userRepo)
	jwtService := auth.NewJWTService(cfg.JWT.Secret, cfg.JWT.ExpiryHours)
	// [GEN-SERVICE]

	// ── HTTP engine ───────────────────────────────────────────────────────────
	app := engine.New(cfg, db, nav)
	app.Register(dashboardhandler.New(nav).AsModule())
	// [GEN-REGISTER]

	// Auth routes (public)
	authH := authhandler.New(userService, jwtService)
	authH.Register(app.Fiber())

	// Protected routes
	protected := app.Fiber().Group("/", middleware.Auth(jwtService))
	userH := userhandler.New(userService, nav)
	userH.Register(protected)

	// ── Graceful shutdown ─────────────────────────────────────────────────────
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("🚀 go-templio listening on :%s (env=%s)", cfg.App.Port, cfg.App.Env)
		if err := app.Run(); err != nil {
			log.Printf("server stopped: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down...")
	shutCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = app.Fiber().ShutdownWithContext(shutCtx)
	log.Println("goodbye.")
}

// seedAdmin creates the default admin account if no users exist.
func seedAdmin(ctx context.Context, repo *userrepo.Repository) {
	_, err := repo.FindByEmail(ctx, "admin@templio.local")
	if err == nil {
		return // already seeded
	}

	hash, _ := bcrypt.GenerateFromPassword([]byte("Admin1234!"), 12)
	u := &user.User{
		Name:     "Admin",
		Email:    "admin@templio.local",
		Password: string(hash),
		Role:     user.RoleAdmin,
		Active:   true,
	}
	if err := repo.Create(ctx, u); err != nil {
		log.Printf("seed admin: %v", err)
		return
	}
	log.Println("🔑 Default admin seeded: admin@templio.local / Admin1234!")
}
