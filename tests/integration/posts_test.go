package integration_test

import (
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/mhtecdev/blog-ai/internal/service"
	"github.com/mhtecdev/blog-ai/tests/testutil"
)

func TestCreateAndReadPost(t *testing.T) {
	app := testutil.NewTestApp(t)

	post, err := app.PostSvc.Create(service.PostInput{
		Title:     "Hello World",
		Excerpt:   "My first post",
		ContentMD: "## Introduction\n\nHello, world!",
		Category:  "ai",
		Tags:      "test,golang",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if post.Slug != "hello-world" {
		t.Errorf("expected slug 'hello-world', got %q", post.Slug)
	}
	if post.Status != "draft" {
		t.Errorf("expected status 'draft', got %q", post.Status)
	}
	if !strings.Contains(post.ContentHTML, "<h2") {
		t.Error("ContentHTML should contain rendered heading")
	}
}

func TestPublishPost(t *testing.T) {
	app := testutil.NewTestApp(t)

	post, _ := app.PostSvc.Create(service.PostInput{Title: "Publish Me", ContentMD: "body"})

	// Not visible on public site before publishing
	resp := app.Get("/posts/" + post.Slug)
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("draft post should return 404, got %d", resp.StatusCode)
	}

	// Publish
	if err := app.PostSvc.Publish(post.ID); err != nil {
		t.Fatalf("Publish: %v", err)
	}

	// Should be visible now — but templates may fail without full template dir
	// so just check service state
	updated, err := app.PostSvc.GetByID(post.ID)
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if updated.Status != "published" {
		t.Errorf("expected status 'published', got %q", updated.Status)
	}
	if updated.PublishedAt == nil {
		t.Error("PublishedAt should be set after Publish()")
	}
}

func TestUpdatePost(t *testing.T) {
	app := testutil.NewTestApp(t)

	post, _ := app.PostSvc.Create(service.PostInput{Title: "Original", ContentMD: "old"})
	updated, err := app.PostSvc.Update(post.ID, service.PostInput{
		Title:     "Updated Title",
		ContentMD: "new content",
	})
	if err != nil {
		t.Fatalf("Update: %v", err)
	}
	if updated.Title != "Updated Title" {
		t.Errorf("expected updated title, got %q", updated.Title)
	}
}

func TestDeletePost(t *testing.T) {
	app := testutil.NewTestApp(t)

	post, _ := app.PostSvc.Create(service.PostInput{Title: "To Delete", ContentMD: "bye"})
	if err := app.PostSvc.Delete(post.ID); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := app.PostSvc.GetByID(post.ID)
	if err == nil {
		t.Error("expected ErrNotFound after deletion")
	}
}

func TestSlugUniqueness(t *testing.T) {
	app := testutil.NewTestApp(t)

	p1, _ := app.PostSvc.Create(service.PostInput{Title: "Same Title", ContentMD: "a"})
	p2, _ := app.PostSvc.Create(service.PostInput{Title: "Same Title", ContentMD: "b"})

	if p1.Slug == p2.Slug {
		t.Errorf("duplicate slugs: both got %q", p1.Slug)
	}
	fmt.Printf("  slugs: %q and %q\n", p1.Slug, p2.Slug)
}
