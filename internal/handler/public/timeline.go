package public

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/service"
)

type TimelineHandler struct {
	posts *service.PostService
}

func NewTimelineHandler(posts *service.PostService) *TimelineHandler {
	return &TimelineHandler{posts: posts}
}

func (h *TimelineHandler) Handle(c *fiber.Ctx) error {
	posts, err := h.posts.ListPublished()
	if err != nil {
		return err
	}
	return c.Render("public/timeline", fiber.Map{
		"Title": "Timeline",
		"Posts": posts,
	}, "layouts/base")
}
