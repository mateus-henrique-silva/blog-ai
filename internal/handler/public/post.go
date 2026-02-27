package public

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/service"
)

type PostHandler struct {
	posts     *service.PostService
	analytics *service.AnalyticsService
}

func NewPostHandler(posts *service.PostService, analytics *service.AnalyticsService) *PostHandler {
	return &PostHandler{posts: posts, analytics: analytics}
}

func (h *PostHandler) Show(c *fiber.Ctx) error {
	slug := c.Params("slug")
	post, err := h.posts.GetBySlug(slug)
	if errors.Is(err, service.ErrNotFound) {
		return c.Status(fiber.StatusNotFound).Render("public/404", fiber.Map{
			"Title": "Post not found",
		}, "layouts/base")
	}
	if err != nil {
		return err
	}

	if !post.IsPublished() {
		return c.Status(fiber.StatusNotFound).Render("public/404", fiber.Map{
			"Title": "Post not found",
		}, "layouts/base")
	}

	// Record view asynchronously
	ip := c.IP()
	ua := string(c.Request().Header.UserAgent())
	h.analytics.RecordView(post.ID, ip, ua)

	return c.Render("public/post", fiber.Map{
		"Title": post.Title,
		"Post":  post,
	}, "layouts/base")
}
