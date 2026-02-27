package studio

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/model"
	"github.com/mhtecdev/blog-ai/internal/service"
)

type PostsHandler struct {
	posts *service.PostService
	media *service.MediaService
}

func NewPostsHandler(posts *service.PostService, media *service.MediaService) *PostsHandler {
	return &PostsHandler{posts: posts, media: media}
}

func (h *PostsHandler) List(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.AdminUser)
	posts, err := h.posts.ListAll()
	if err != nil {
		return err
	}
	return c.Render("studio/posts_list", fiber.Map{
		"Title":   "All Posts",
		"Section": "posts",
		"User":    user,
		"Posts":   posts,
	}, "layouts/studio")
}

func (h *PostsHandler) New(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.AdminUser)
	return c.Render("studio/post_editor", fiber.Map{
		"Title":      "New Post",
		"Section":    "posts",
		"User":       user,
		"Post":       nil,
		"LoadEditor": true,
	}, "layouts/studio")
}

func (h *PostsHandler) Create(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.AdminUser)

	input := service.PostInput{
		Title:      c.FormValue("title"),
		Excerpt:    c.FormValue("excerpt"),
		ContentMD:  c.FormValue("content_md"),
		CoverImage: c.FormValue("cover_image"),
		Category:   c.FormValue("category"),
		Tags:       c.FormValue("tags"),
	}

	if input.Title == "" {
		return c.Status(fiber.StatusUnprocessableEntity).Render("studio/post_editor", fiber.Map{
			"Title":      "New Post",
			"Section":    "posts",
			"User":       user,
			"Error":      "Title is required.",
			"Input":      input,
			"LoadEditor": true,
		}, "layouts/studio")
	}

	post, err := h.posts.Create(input)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).Render("studio/post_editor", fiber.Map{
			"Title":      "New Post",
			"Section":    "posts",
			"User":       user,
			"Error":      "Failed to create post.",
			"Input":      input,
			"LoadEditor": true,
		}, "layouts/studio")
	}

	return c.Redirect("/studio/posts/"+strconv.FormatInt(post.ID, 10)+"/edit?created=1", fiber.StatusSeeOther)
}

func (h *PostsHandler) Edit(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.AdminUser)
	id, err := parseID(c)
	if err != nil {
		return fiber.ErrBadRequest
	}

	post, err := h.posts.GetByID(id)
	if errors.Is(err, service.ErrNotFound) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return err
	}

	flash := ""
	if c.Query("created") == "1" {
		flash = "Post created successfully."
	}
	if c.Query("saved") == "1" {
		flash = "Post saved."
	}

	return c.Render("studio/post_editor", fiber.Map{
		"Title":      "Edit Post",
		"Section":    "posts",
		"User":       user,
		"Post":       post,
		"Flash":      flash,
		"LoadEditor": true,
	}, "layouts/studio")
}

func (h *PostsHandler) Update(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.AdminUser)
	id, err := parseID(c)
	if err != nil {
		return fiber.ErrBadRequest
	}

	input := service.PostInput{
		Title:      c.FormValue("title"),
		Excerpt:    c.FormValue("excerpt"),
		ContentMD:  c.FormValue("content_md"),
		CoverImage: c.FormValue("cover_image"),
		Category:   c.FormValue("category"),
		Tags:       c.FormValue("tags"),
	}

	if input.Title == "" {
		post, _ := h.posts.GetByID(id)
		return c.Status(fiber.StatusUnprocessableEntity).Render("studio/post_editor", fiber.Map{
			"Title":      "Edit Post",
			"Section":    "posts",
			"User":       user,
			"Post":       post,
			"Error":      "Title is required.",
			"LoadEditor": true,
		}, "layouts/studio")
	}

	_, err = h.posts.Update(id, input)
	if errors.Is(err, service.ErrNotFound) {
		return fiber.ErrNotFound
	}
	if err != nil {
		return err
	}

	return c.Redirect("/studio/posts/"+strconv.FormatInt(id, 10)+"/edit?saved=1", fiber.StatusSeeOther)
}

func (h *PostsHandler) Delete(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.posts.Delete(id); err != nil && !errors.Is(err, service.ErrNotFound) {
		return err
	}
	return c.Redirect("/studio/posts", fiber.StatusSeeOther)
}

func (h *PostsHandler) Publish(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.posts.Publish(id); err != nil && !errors.Is(err, service.ErrNotFound) {
		return err
	}
	return c.Redirect("/studio/posts/"+strconv.FormatInt(id, 10)+"/edit", fiber.StatusSeeOther)
}

func (h *PostsHandler) Unpublish(c *fiber.Ctx) error {
	id, err := parseID(c)
	if err != nil {
		return fiber.ErrBadRequest
	}
	if err := h.posts.Unpublish(id); err != nil && !errors.Is(err, service.ErrNotFound) {
		return err
	}
	return c.Redirect("/studio/posts/"+strconv.FormatInt(id, 10)+"/edit", fiber.StatusSeeOther)
}

func (h *PostsHandler) Upload(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.AdminUser)

	fh, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "no file provided"})
	}

	media, err := h.media.Upload(fh, user.ID)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{
		"url":       media.URL,
		"mime_type": media.MimeType,
		"filename":  media.Original,
	})
}

func parseID(c *fiber.Ctx) (int64, error) {
	return strconv.ParseInt(c.Params("id"), 10, 64)
}
