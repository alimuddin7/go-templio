// Package userhandler wires Fiber routes to the user.Service.
// It owns NO business logic — it only translates HTTP ↔ domain.
package userhandler

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"

	"github.com/alimuddin7/go-templio/internal/domain/user"
	"github.com/alimuddin7/go-templio/internal/engine"
	"github.com/alimuddin7/go-templio/internal/navigation"
	userviews "github.com/alimuddin7/go-templio/views/user"
)

// Handler holds dependencies for the user HTTP layer.
type Handler struct {
	svc user.Service
	nav *navigation.Registry
}

// New creates a Handler. Accepts the interface — testable without a real DB.
func New(svc user.Service, nav *navigation.Registry) *Handler {
	return &Handler{svc: svc, nav: nav}
}

// Register mounts all user routes under the given router group.
func (h *Handler) Register(r fiber.Router) {
	g := r.Group("/users")
	g.Get("/", h.list)
	g.Get("/new", h.createForm)
	g.Post("/", h.create)
	g.Get("/:id/edit", h.updateForm)
	g.Post("/:id", h.update) // HTML forms can't send PUT
	g.Post("/:id/delete", h.delete)
}

// AsModule returns a plugin function for the engine.
func (h *Handler) AsModule() engine.Module {
	return func(r fiber.Router, nav *navigation.Registry) {
		h.Register(r)
	}
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (h *Handler) list(c fiber.Ctx) error {
	search := c.Query("q")
	page := queryInt(c, "p", 1)
	pageSize := queryInt(c, "ps", 20)

	users, total, err := h.svc.List(c.Context(), search, page, pageSize)
	if err != nil {
		return err
	}

	return engine.Render(c, userviews.List(users, total, page, pageSize, search, h.nav.Items()))
}

func (h *Handler) createForm(c fiber.Ctx) error {
	return engine.Render(c, userviews.Create(h.nav.Items(), nil))
}

func (h *Handler) create(c fiber.Ctx) error {
	var in user.CreateUserInput
	if err := c.Bind().Body(&in); err != nil {
		return engine.Render(c, userviews.Create(h.nav.Items(), err), fiber.StatusUnprocessableEntity)
	}

	if _, err := h.svc.Create(c.Context(), in); err != nil {
		return engine.Render(c, userviews.Create(h.nav.Items(), err), fiber.StatusUnprocessableEntity)
	}

	return c.Redirect().To("/users")
}


func (h *Handler) updateForm(c fiber.Ctx) error {
	id, err := pathID(c)
	if err != nil {
		return fiber.ErrBadRequest
	}

	u, err := h.svc.GetByID(c.Context(), id)
	if errors.Is(err, user.ErrNotFound) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return err
	}

	return engine.Render(c, userviews.Update(u, h.nav.Items(), nil))
}

func (h *Handler) update(c fiber.Ctx) error {
	id, err := pathID(c)
	if err != nil {
		return fiber.ErrBadRequest
	}

	var in user.UpdateUserInput
	if err := c.Bind().Body(&in); err != nil {
		u, _ := h.svc.GetByID(c.Context(), id)
		return engine.Render(c, userviews.Update(u, h.nav.Items(), err), fiber.StatusUnprocessableEntity)
	}

	if _, err := h.svc.Update(c.Context(), id, in); err != nil {
		u, _ := h.svc.GetByID(c.Context(), id)
		return engine.Render(c, userviews.Update(u, h.nav.Items(), err), fiber.StatusUnprocessableEntity)
	}

	return c.Redirect().To("/users")
}

func (h *Handler) delete(c fiber.Ctx) error {
	id, err := pathID(c)
	if err != nil {
		return fiber.ErrBadRequest
	}

	if err := h.svc.Delete(c.Context(), id); err != nil {
		return err
	}

	return c.Redirect().To("/users")
}

// ── helpers ──────────────────────────────────────────────────────────────────

func pathID(c fiber.Ctx) (int64, error) {
	return strconv.ParseInt(c.Params("id"), 10, 64)
}

func queryInt(c fiber.Ctx, key string, def int) int {
	v := c.Query(key)
	if v == "" {
		return def
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return def
	}
	return n
}
