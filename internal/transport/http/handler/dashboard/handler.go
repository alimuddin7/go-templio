package dashboardhandler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/alimuddin7/go-templio/internal/engine"
	"github.com/alimuddin7/go-templio/internal/navigation"
	"github.com/alimuddin7/go-templio/views/dashboard"
	"github.com/alimuddin7/go-templio/views/layout"
)

type Handler struct {
	nav *navigation.Registry
}

func New(nav *navigation.Registry) *Handler {
	return &Handler{nav: nav}
}

func (h *Handler) Register(r fiber.Router) {
	r.Get("/", h.index)
}

func (h *Handler) AsModule() engine.Module {
	return func(r fiber.Router, nav *navigation.Registry) {
		h.Register(r)
	}
}

func (h *Handler) index(c fiber.Ctx) error {
	name, _ := c.Locals("name").(string)
	role, _ := c.Locals("role").(string)

	return engine.Render(c, dashboard.Index(layout.PageData{
		Title:       "Dashboard",
		NavItems:    h.nav.Items(),
		CurrentPath: "/",
		UserName:    name,
		UserRole:    role,
	}))
}
