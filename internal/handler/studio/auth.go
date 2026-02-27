package studio

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/config"
	"github.com/mhtecdev/blog-ai/internal/middleware"
	"github.com/mhtecdev/blog-ai/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
	cfg  *config.Config
}

func NewAuthHandler(auth *service.AuthService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{auth: auth, cfg: cfg}
}

func (h *AuthHandler) ShowLogin(c *fiber.Ctx) error {
	// Already logged in? Redirect to dashboard.
	if sid := c.Cookies(middleware.SessionCookieName); sid != "" {
		if _, err := h.auth.Validate(sid); err == nil {
			return c.Redirect("/studio/dashboard", fiber.StatusSeeOther)
		}
	}
	return c.Render("studio/login", fiber.Map{
		"Title": "Sign in",
	})
}

func (h *AuthHandler) ProcessLogin(c *fiber.Ctx) error {
	username := c.FormValue("username")
	password := c.FormValue("password")

	session, err := h.auth.Login(username, password, c.IP(), string(c.Request().Header.UserAgent()))
	if err != nil {
		errMsg := "Invalid username or password."
		if !errors.Is(err, service.ErrInvalidCredentials) {
			errMsg = "An error occurred. Please try again."
		}
		return c.Status(fiber.StatusUnauthorized).Render("studio/login", fiber.Map{
			"Title": "Sign in",
			"Error": errMsg,
		})
	}

	secure := h.cfg.AppEnv == "production"
	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		Secure:   secure,
		SameSite: "Lax",
		Path:     "/",
	})

	return c.Redirect("/studio/dashboard", fiber.StatusSeeOther)
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sid := c.Cookies(middleware.SessionCookieName)
	if sid != "" {
		_ = h.auth.Logout(sid)
	}
	c.Cookie(&fiber.Cookie{
		Name:    middleware.SessionCookieName,
		Value:   "",
		Expires: time.Unix(0, 0),
		Path:    "/",
	})
	return c.Redirect("/studio/login", fiber.StatusSeeOther)
}
