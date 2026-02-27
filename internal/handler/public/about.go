package public

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/model"
)

func AboutHandler(c *fiber.Ctx) error {
	return c.Render("public/about", fiber.Map{
		"Title":  "About",
		"Author": model.Author,
	}, "layouts/base")
}
