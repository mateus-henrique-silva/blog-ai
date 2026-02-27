package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/service"
)

// RecordView records a page view for a post asynchronously.
// It expects the post ID to be set in c.Locals("postID") by the handler before calling Next.
// Usage: apply after the handler has already resolved the post.
//
// Because Fiber handlers run sequentially, this middleware is applied as a
// post-processing step by calling it from within the public post handler.
func RecordViewMiddleware(analyticsSvc *service.AnalyticsService) func(postID int64, c *fiber.Ctx) {
	return func(postID int64, c *fiber.Ctx) {
		ip := c.IP()
		ua := string(c.Request().Header.UserAgent())
		go analyticsSvc.RecordView(postID, ip, ua)
	}
}
