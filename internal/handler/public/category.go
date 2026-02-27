package public

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/service"
)

type CategoryHandler struct {
	posts *service.PostService
}

func NewCategoryHandler(posts *service.PostService) *CategoryHandler {
	return &CategoryHandler{posts: posts}
}

func (h *CategoryHandler) List(c *fiber.Ctx) error {
	categories, err := h.posts.ListCategories()
	if err != nil {
		return err
	}
	return c.Render("public/categories", fiber.Map{
		"Title":      "Categories",
		"Categories": categories,
	}, "layouts/base")
}

func (h *CategoryHandler) Show(c *fiber.Ctx) error {
	slug := c.Params("slug")
	posts, err := h.posts.ListPublishedByCategory(slug)
	if err != nil {
		return err
	}
	return c.Render("public/category_detail", fiber.Map{
		"Title":    slug,
		"Category": slug,
		"Posts":    posts,
	}, "layouts/base")
}
