// Package authhandler provides login/logout HTTP handlers.
package authhandler

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"golang.org/x/crypto/bcrypt"

	"templio.local/cms/internal/auth"
	"templio.local/cms/internal/domain/user"
	"templio.local/cms/internal/engine"
	authviews "templio.local/cms/views/auth"
)

const cookieName = "templio_token"

// Handler provides login/logout routes.
type Handler struct {
	userSvc user.Service
	jwt     *auth.JWTService
}

// New creates an authhandler Handler.
func New(userSvc user.Service, jwtSvc *auth.JWTService) *Handler {
	return &Handler{userSvc: userSvc, jwt: jwtSvc}
}

// Register mounts auth routes on the given router.
func (h *Handler) Register(r fiber.Router) {
	r.Get("/login", h.loginForm)
	r.Post("/login", h.login)
	r.Get("/logout", h.logout)
}

// ── Handlers ──────────────────────────────────────────────────────────────────

func (h *Handler) loginForm(c fiber.Ctx) error {
	return engine.Render(c, authviews.Login("", nil))
}

type loginInput struct {
	Email    string `form:"email"`
	Password string `form:"password"`
}

func (h *Handler) login(c fiber.Ctx) error {
	var in loginInput
	if err := c.Bind().Body(&in); err != nil {
		return engine.Render(c, authviews.Login(in.Email, err), fiber.StatusUnprocessableEntity)
	}

	u, err := h.userSvc.GetByEmail(c.Context(), in.Email)
	if errors.Is(err, user.ErrNotFound) {
		return engine.Render(c, authviews.Login(in.Email, errors.New("invalid email or password")), fiber.StatusUnauthorized)
	}
	if err != nil {
		return err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(in.Password)); err != nil {
		return engine.Render(c, authviews.Login(in.Email, errors.New("invalid email or password")), fiber.StatusUnauthorized)
	}

	pair, err := h.jwt.Issue(u.ID, u.Name, u.Email, string(u.Role))
	if err != nil {
		return err
	}

	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    pair.AccessToken,
		Expires:  pair.ExpiresAt,
		HTTPOnly: true,
		Secure:   false, // set true behind TLS
		SameSite: "Lax",
	})

	return c.Redirect().To("/")
}

func (h *Handler) logout(c fiber.Ctx) error {
	c.ClearCookie(cookieName)
	return c.Redirect().To("/login")
}
