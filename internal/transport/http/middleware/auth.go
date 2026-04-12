// Package middleware provides reusable Fiber middleware constructors.
// All middleware is created via constructor functions — no global state.
package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v3"

	"github.com/alimuddin7/go-templio/internal/auth"
)

const (
	cookieName    = "templio_token"
	contextKeyUID  = "uid"
	contextKeyName = "name"
	contextKeyRole = "role"
)

// Auth returns a Fiber middleware that validates the JWT from the cookie.
// Unauthenticated requests are redirected to /login.
func Auth(jwtSvc *auth.JWTService) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := extractToken(c)
		if token == "" {
			return c.Redirect().To("/login")
		}

		claims, err := jwtSvc.Validate(token)
		if err != nil {
			c.ClearCookie(cookieName)
			return c.Redirect().To("/login")
		}

		// Store claims in context locals for downstream handlers.
		c.Locals(contextKeyUID, claims.UserID)
		c.Locals(contextKeyName, claims.Name)
		c.Locals(contextKeyRole, claims.Role)

		return c.Next()
	}
}

// RequireRole returns a middleware that enforces a minimum role.
// Call after Auth middleware.
func RequireRole(role string) fiber.Handler {
	return func(c fiber.Ctx) error {
		userRole, ok := c.Locals(contextKeyRole).(string)
		if !ok || userRole != role {
			return fiber.ErrForbidden
		}
		return c.Next()
	}
}

// Logger returns a simple request logger middleware.
func Logger() fiber.Handler {
	return func(c fiber.Ctx) error {
		err := c.Next()
		// Structured log: method, path, status, latency.
		// In production, replace with zerolog/slog.
		_ = err
		return err
	}
}

// ── helpers ──────────────────────────────────────────────────────────────────

func extractToken(c fiber.Ctx) string {
	// 1. Cookie (browser)
	if t := c.Cookies(cookieName); t != "" {
		return t
	}
	// 2. Bearer header (API clients)
	auth := c.Get(fiber.HeaderAuthorization)
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}
	return ""
}
