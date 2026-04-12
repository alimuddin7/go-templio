// Package engine provides the central bootstrap for the Fiber HTTP server
// and all infrastructure wiring. It owns NO business logic.
package engine

import (
	"errors"
	"fmt"

	"github.com/a-h/templ"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/templui/templui/utils"

	"templio.local/cms/internal/config"
	"templio.local/cms/internal/navigation"
)

// Module is a function that registers routes and navigation items
// into the Fiber router and the navigation registry.
// This is the "plugin" contract — the engine never imports a concrete module.
type Module func(router fiber.Router, nav *navigation.Registry)

// App is the central application container.
// Call New() then Register() for each module, then Run().
type App struct {
	fiber  *fiber.App
	cfg    *config.Config
	nav    *navigation.Registry
	db     *Database
}

// New builds the Fiber instance and applies base middleware.
// It does NOT start listening — call Run() for that.
func New(cfg *config.Config, db *Database, nav *navigation.Registry) *App {
	fiberCfg := fiber.Config{
		AppName:               "go-templio",
		ErrorHandler:          errorHandler,
		StreamRequestBody:     true,
	}

	f := fiber.New(fiberCfg)

	// ── Global middleware stack ───────────────────────────────────────────────
	f.Use(recover.New())
	f.Use(requestid.New())
	f.Use(cors.New())

	// Configure templui JS component scripts to be served from static directory
	utils.ScriptURL = func(path string) string {
		// path is like "/templui/js/selectbox.min.js" → remap to /static/js/templui/selectbox.min.js
		baseFile := path[len("/templui/js/"):]
		return "/static/js/templui/" + baseFile
	}

	// Serve static assets (CSS, JS, icons)
	f.Get("/static/*", static.New("./static"))

	return &App{
		fiber: f,
		cfg:   cfg,
		nav:   nav,
		db:    db,
	}
}

// Register applies a Module to the router and navigation registry.
// Modules are plain functions — the engine stays untouched.
func (a *App) Register(modules ...Module) {
	for _, m := range modules {
		m(a.fiber, a.nav)
	}
}

// Run starts the Fiber HTTP server.
func (a *App) Run() error {
	addr := fmt.Sprintf(":%s", a.cfg.App.Port)
	return a.fiber.Listen(addr, fiber.ListenConfig{
		EnablePrefork: a.cfg.App.Prefork,
	})
}

// Fiber exposes the underlying *fiber.App for testing or advanced configuration.
func (a *App) Fiber() *fiber.App { return a.fiber }

// Nav exposes the navigation registry (e.g. to pass into Templ layouts).
func (a *App) Nav() *navigation.Registry { return a.nav }

// ── Templ rendering helper ───────────────────────────────────────────────────

// Render is a convenience function that renders a templ.Component into a Fiber response.
func Render(c fiber.Ctx, component templ.Component, status ...int) error {
	code := fiber.StatusOK
	if len(status) > 0 {
		code = status[0]
	}
	c.Set(fiber.HeaderContentType, fiber.MIMETextHTMLCharsetUTF8)
	c.Status(code)
	return component.Render(c.Context(), c.Response().BodyWriter())
}

// ── Error handler ────────────────────────────────────────────────────────────

func errorHandler(c fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	var fe *fiber.Error
	if ok := errors.As(err, &fe); ok {
		code = fe.Code
	}
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
