package middleware

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/service"
)

const SessionCookieName = "session_id"

// RequireAuth validates the session cookie and sets "user" in locals.
// Unauthenticated requests are redirected to /studio/login.
func RequireAuth(authSvc *service.AuthService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies(SessionCookieName)
		if sessionID == "" {
			return c.Redirect("/studio/login", fiber.StatusSeeOther)
		}

		user, err := authSvc.Validate(sessionID)
		if err != nil {
			if errors.Is(err, service.ErrSessionExpired) {
				c.ClearCookie(SessionCookieName)
			}
			return c.Redirect("/studio/login", fiber.StatusSeeOther)
		}

		c.Locals("user", user)
		return c.Next()
	}
}
