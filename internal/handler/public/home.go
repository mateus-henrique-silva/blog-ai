package public

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/service"
)

type HomeHandler struct {
	posts *service.PostService
}

func NewHomeHandler(posts *service.PostService) *HomeHandler {
	return &HomeHandler{posts: posts}
}

func (h *HomeHandler) Handle(c *fiber.Ctx) error {
	posts, err := h.posts.ListPublished()
	if err != nil {
		return err
	}

	// Featured = first post; recent = next N
	var featured interface{}
	recent := posts
	if len(posts) > 0 {
		featured = posts[0]
		recent = posts[1:]
	}

	categories, _ := h.posts.ListCategories()

	return c.Render("public/home", fiber.Map{
		"Title":      "AI Studies",
		"Posts":      recent,
		"Featured":   featured,
		"Categories": categories,
	}, "layouts/base")
}
