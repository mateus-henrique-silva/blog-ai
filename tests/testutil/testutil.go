package testutil

import (
	"bytes"
	"encoding/json"
	"html/template"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	htmlEngine "github.com/gofiber/template/html/v2"
	"github.com/mhtecdev/blog-ai/internal/config"
	"github.com/mhtecdev/blog-ai/internal/database"
	handlerPublic "github.com/mhtecdev/blog-ai/internal/handler/public"
	handlerStudio "github.com/mhtecdev/blog-ai/internal/handler/studio"
	"github.com/mhtecdev/blog-ai/internal/middleware"
	"github.com/mhtecdev/blog-ai/internal/repository"
	"github.com/mhtecdev/blog-ai/internal/service"
)

// TestApp wraps a Fiber app and exposes helpers for testing.
type TestApp struct {
	App     *fiber.App
	AuthSvc *service.AuthService
	PostSvc *service.PostService
}

func NewTestApp(t *testing.T) *TestApp {
	t.Helper()

	cfg := &config.Config{
		AppEnv:          "development",
		AppPort:         "3000",
		AppSecret:       "test-secret-key-32-bytes-xxxxxxxx",
		IPHashSecret:    "test-ip-hash-secret-for-tests",
		DBPath:          ":memory:",
		UploadDir:       t.TempDir(),
		UploadMaxMB:     5,
		SessionDuration: 1 * time.Hour,
		RateLimitLogin:  3,
		RateLimitWindow: 5 * time.Second,
		CSPMode:         "lenient",
	}

	db := database.Open(":memory:?_foreign_keys=on")
	database.RunMigrations(db)

	userRepo      := repository.NewUserRepo(db)
	postRepo      := repository.NewPostRepo(db)
	sessionRepo   := repository.NewSessionRepo(db)
	analyticsRepo := repository.NewAnalyticsRepo(db)
	mediaRepo     := repository.NewMediaRepo(db)

	authSvc, err := service.NewAuthService(userRepo, sessionRepo, cfg)
	if err != nil {
		t.Fatalf("testutil: failed to create auth service: %v", err)
	}
	postSvc      := service.NewPostService(postRepo)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo, cfg)
	mediaSvc     := service.NewMediaService(mediaRepo, cfg)

	// Use a minimal inline template engine for tests
	engine := htmlEngine.New("../../web/templates", ".html")
	engine.AddFunc("safeHTML", func(s string) template.HTML { return template.HTML(s) })
	engine.AddFunc("inc", func(i int) int { return i + 1 })
	engine.AddFunc("jsonViews", func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	})

	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: func(c *fiber.Ctx, err error) error { return c.Status(500).SendString(err.Error()) },
	})

	app.Static("/static", "../../web/static")

	rateLimiter := middleware.NewRateLimiter(cfg)
	go rateLimiter.Cleanup()
	authMW := middleware.RequireAuth(authSvc)

	app.Use(middleware.SecurityHeaders(cfg))

	// Public routes
	homeH     := handlerPublic.NewHomeHandler(postSvc)
	postH     := handlerPublic.NewPostHandler(postSvc, analyticsSvc)
	categoryH := handlerPublic.NewCategoryHandler(postSvc)
	timelineH := handlerPublic.NewTimelineHandler(postSvc)

	app.Get("/", homeH.Handle)
	app.Get("/posts/:slug", postH.Show)
	app.Get("/categories", categoryH.List)
	app.Get("/categories/:slug", categoryH.Show)
	app.Get("/timeline", timelineH.Handle)

	// Studio routes
	authH      := handlerStudio.NewAuthHandler(authSvc, cfg)
	dashboardH := handlerStudio.NewDashboardHandler(postSvc, analyticsSvc)
	postsH     := handlerStudio.NewPostsHandler(postSvc, mediaSvc)
	metricsH   := handlerStudio.NewMetricsHandler(analyticsSvc)

	studio := app.Group("/studio")
	studio.Get("/login", authH.ShowLogin)
	studio.Post("/login", rateLimiter.Middleware(), authH.ProcessLogin)
	studio.Post("/logout", authMW, authH.Logout)
	studio.Get("/", func(c *fiber.Ctx) error { return c.Redirect("/studio/dashboard", fiber.StatusSeeOther) })
	studio.Get("/dashboard", authMW, dashboardH.Handle)
	studio.Get("/posts", authMW, postsH.List)
	studio.Get("/posts/new", authMW, postsH.New)
	studio.Post("/posts", authMW, postsH.Create)
	studio.Get("/posts/:id/edit", authMW, postsH.Edit)
	studio.Post("/posts/:id", authMW, postsH.Update)
	studio.Post("/posts/:id/delete", authMW, postsH.Delete)
	studio.Post("/posts/:id/publish", authMW, postsH.Publish)
	studio.Post("/posts/:id/unpublish", authMW, postsH.Unpublish)
	studio.Post("/upload", authMW, postsH.Upload)
	studio.Get("/metrics", authMW, metricsH.Handle)

	return &TestApp{App: app, AuthSvc: authSvc, PostSvc: postSvc}
}

// Do performs a test HTTP request.
func (ta *TestApp) Do(method, path string, body io.Reader, headers map[string]string) *http.Response {
	req := httptest.NewRequest(method, path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, _ := ta.App.Test(req, -1)
	return resp
}

// Get is a convenience wrapper.
func (ta *TestApp) Get(path string) *http.Response {
	return ta.Do(http.MethodGet, path, nil, nil)
}

// PostForm submits form data.
func (ta *TestApp) PostForm(path string, form map[string]string, cookies []*http.Cookie) *http.Response {
	vals := make([]string, 0, len(form))
	for k, v := range form {
		vals = append(vals, k+"="+v)
	}
	body := strings.NewReader(strings.Join(vals, "&"))
	req := httptest.NewRequest(http.MethodPost, path, body)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for _, c := range cookies {
		req.AddCookie(c)
	}
	resp, _ := ta.App.Test(req, -1)
	return resp
}

// PostJSON sends a JSON body.
func (ta *TestApp) PostJSON(path string, v interface{}) *http.Response {
	b, _ := json.Marshal(v)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := ta.App.Test(req, -1)
	return resp
}

// SeedUser creates a test admin user and returns the session cookie.
func (ta *TestApp) SeedUser(t *testing.T, username, password string) *http.Cookie {
	t.Helper()
	if err := ta.AuthSvc.CreateUser(username, password, ""); err != nil {
		t.Fatalf("SeedUser: %v", err)
	}
	session, err := ta.AuthSvc.Login(username, password, "127.0.0.1", "test-agent")
	if err != nil {
		t.Fatalf("SeedUser login: %v", err)
	}
	return &http.Cookie{Name: "session_id", Value: session.ID}
}
