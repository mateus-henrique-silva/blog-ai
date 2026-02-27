package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/config"
)

const (
	cspStrict = "default-src 'self'; " +
		"script-src 'self'; " +
		"style-src 'self' 'unsafe-inline'; " +
		"img-src 'self' data: blob:; " +
		"media-src 'self'; " +
		"font-src 'self'; " +
		"connect-src 'self'; " +
		"frame-ancestors 'none'; " +
		"base-uri 'self'; " +
		"form-action 'self'"

	cspLenient = "default-src 'self' 'unsafe-inline' 'unsafe-eval'; " +
		"img-src * data: blob:; " +
		"media-src *; " +
		"connect-src *; " +
		"frame-ancestors 'none'"
)

func SecurityHeaders(cfg *config.Config) fiber.Handler {
	csp := cspStrict
	if cfg.CSPMode == "lenient" {
		csp = cspLenient
	}

	return func(c *fiber.Ctx) error {
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY")
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		c.Set("Content-Security-Policy", csp)

		if cfg.AppEnv == "production" {
			c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}
		return c.Next()
	}
}
