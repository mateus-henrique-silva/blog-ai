package service

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/microcosm-cc/bluemonday"
	"github.com/mhtecdev/blog-ai/internal/model"
	"github.com/mhtecdev/blog-ai/internal/repository"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/parser"
	"github.com/yuin/goldmark/renderer/html"
)

var (
	ErrSlugConflict = errors.New("slug already in use")
	ErrNotFound     = errors.New("not found")
)

type PostInput struct {
	Title      string
	Excerpt    string
	ContentMD  string
	CoverImage string
	Category   string
	Tags       string
}

type PostService struct {
	repo   *repository.PostRepo
	mdParser goldmark.Markdown
	sanitizer *bluemonday.Policy
}

func NewPostService(repo *repository.PostRepo) *PostService {
	md := goldmark.New(
		goldmark.WithExtensions(
			extension.GFM,
			extension.Footnote,
		),
		goldmark.WithParserOptions(
			parser.WithAutoHeadingID(),
		),
		goldmark.WithRendererOptions(
			html.WithHardWraps(),
			html.WithXHTML(),
			html.WithUnsafe(), // allow raw HTML in Markdown (needed for <video>/<audio>)
		),
	)

	policy := bluemonday.UGCPolicy()
	// Allow video/audio tags inserted by the editor
	policy.AllowElements("video", "audio", "source")
	policy.AllowAttrs("controls", "src", "type", "width", "height").OnElements("video", "audio")
	policy.AllowAttrs("src", "type").OnElements("source")

	return &PostService{repo: repo, mdParser: md, sanitizer: policy}
}

func (s *PostService) GetBySlug(slug string) (*model.Post, error) {
	post, err := s.repo.GetBySlug(slug)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return post, err
}

func (s *PostService) GetByID(id int64) (*model.Post, error) {
	post, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	return post, err
}

func (s *PostService) ListPublished() ([]*model.Post, error) {
	return s.repo.ListPublished()
}

func (s *PostService) ListAll() ([]*model.Post, error) {
	return s.repo.ListAll()
}

func (s *PostService) ListPublishedByCategory(category string) ([]*model.Post, error) {
	return s.repo.ListPublishedByCategory(category)
}

func (s *PostService) ListCategories() ([]string, error) {
	return s.repo.ListCategories()
}

func (s *PostService) Create(input PostInput) (*model.Post, error) {
	slug, err := s.generateSlug(input.Title, 0)
	if err != nil {
		return nil, err
	}

	html := s.renderMarkdown(input.ContentMD)

	post := &model.Post{
		Title:       input.Title,
		Slug:        slug,
		Excerpt:     input.Excerpt,
		ContentMD:   input.ContentMD,
		ContentHTML: html,
		CoverImage:  input.CoverImage,
		Category:    input.Category,
		Tags:        input.Tags,
		Status:      "draft",
	}

	return s.repo.Create(post)
}

func (s *PostService) Update(id int64, input PostInput) (*model.Post, error) {
	existing, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	// Re-generate slug only if title changed
	slug := existing.Slug
	if !strings.EqualFold(existing.Title, input.Title) {
		slug, err = s.generateSlugExcluding(input.Title, existing.Slug)
		if err != nil {
			return nil, err
		}
	}

	html := s.renderMarkdown(input.ContentMD)

	existing.Title = input.Title
	existing.Slug = slug
	existing.Excerpt = input.Excerpt
	existing.ContentMD = input.ContentMD
	existing.ContentHTML = html
	existing.CoverImage = input.CoverImage
	existing.Category = input.Category
	existing.Tags = input.Tags

	return s.repo.Update(existing)
}

func (s *PostService) Publish(id int64) error {
	post, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	now := time.Now()
	post.Status = "published"
	if post.PublishedAt == nil {
		post.PublishedAt = &now
	}
	_, err = s.repo.Update(post)
	return err
}

func (s *PostService) Unpublish(id int64) error {
	post, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	post.Status = "draft"
	_, err = s.repo.Update(post)
	return err
}

func (s *PostService) Delete(id int64) error {
	_, err := s.repo.GetByID(id)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return err
	}
	return s.repo.Delete(id)
}

func (s *PostService) renderMarkdown(md string) string {
	var buf strings.Builder
	if err := s.mdParser.Convert([]byte(md), &buf); err != nil {
		return ""
	}
	return s.sanitizer.Sanitize(buf.String())
}

var nonAlphanumRe = regexp.MustCompile(`[^a-z0-9]+`)

func slugify(title string) string {
	s := strings.ToLower(title)
	var b strings.Builder
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		} else {
			b.WriteRune('-')
		}
	}
	slug := nonAlphanumRe.ReplaceAllString(b.String(), "-")
	slug = strings.Trim(slug, "-")
	if len(slug) > 100 {
		slug = slug[:100]
		slug = strings.TrimRight(slug, "-")
	}
	return slug
}

func (s *PostService) generateSlug(title string, attempt int) (string, error) {
	base := slugify(title)
	slug := base
	if attempt > 0 {
		slug = fmt.Sprintf("%s-%d", base, attempt+1)
	}
	exists, err := s.repo.SlugExists(slug)
	if err != nil {
		return "", err
	}
	if !exists {
		return slug, nil
	}
	return s.generateSlug(title, attempt+1)
}

// generateSlugExcluding generates a slug excluding the currentSlug from collision check.
func (s *PostService) generateSlugExcluding(title, currentSlug string) (string, error) {
	candidate := slugify(title)
	if candidate == currentSlug {
		return currentSlug, nil
	}
	return s.generateSlug(title, 0)
}
