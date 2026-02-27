package main

import (
	"encoding/json"
	"html/template"
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	htmlEngine "github.com/gofiber/template/html/v2"
	"github.com/joho/godotenv"

	"github.com/mhtecdev/blog-ai/internal/config"
	"github.com/mhtecdev/blog-ai/internal/database"
	handlerPublic "github.com/mhtecdev/blog-ai/internal/handler/public"
	handlerStudio "github.com/mhtecdev/blog-ai/internal/handler/studio"
	"github.com/mhtecdev/blog-ai/internal/middleware"
	"github.com/mhtecdev/blog-ai/internal/repository"
	"github.com/mhtecdev/blog-ai/internal/service"
)

func main() {
	// Load .env (no-op if absent — production uses real env vars)
	_ = godotenv.Load()

	cfg := config.Load()

	db := database.Open(cfg.DBPath)
	database.RunMigrations(db)

	// Repositories
	userRepo      := repository.NewUserRepo(db)
	postRepo      := repository.NewPostRepo(db)
	sessionRepo   := repository.NewSessionRepo(db)
	analyticsRepo := repository.NewAnalyticsRepo(db)
	mediaRepo     := repository.NewMediaRepo(db)

	// Services
	authSvc, err := service.NewAuthService(userRepo, sessionRepo, cfg)
	if err != nil {
		log.Fatalf("failed to init auth service: %v", err)
	}
	postSvc      := service.NewPostService(postRepo)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo, cfg)
	mediaSvc     := service.NewMediaService(mediaRepo, cfg)

	// Template engine
	engine := htmlEngine.New("./web/templates", ".html")
	if cfg.IsDevelopment() {
		engine.Reload(true) // hot-reload templates in dev
	}
	// Register template functions
	engine.AddFunc("safeHTML", func(s string) template.HTML {
		return template.HTML(s)
	})
	engine.AddFunc("inc", func(i int) int { return i + 1 })
	engine.AddFunc("jsonViews", func(v interface{}) string {
		b, _ := json.Marshal(v)
		return string(b)
	})

	app := fiber.New(fiber.Config{
		Views:        engine,
		ErrorHandler: errorHandler,
		BodyLimit:    int(cfg.UploadMaxMB) * 1024 * 1024,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	})

	// Global middleware
	app.Use(recover.New())
	app.Use(middleware.SecurityHeaders(cfg))
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} ${method} ${path} ${latency}\n",
	}))

	// Static files
	app.Static("/static", "./web/static")

	// Rate limiter for login endpoint
	rateLimiter := middleware.NewRateLimiter(cfg)
	go rateLimiter.Cleanup()

	// Auth middleware for protected routes
	authMW := middleware.RequireAuth(authSvc)

	// ─── Public routes ───────────────────────────────────────────────────────
	homeH     := handlerPublic.NewHomeHandler(postSvc)
	postH     := handlerPublic.NewPostHandler(postSvc, analyticsSvc)
	categoryH := handlerPublic.NewCategoryHandler(postSvc)
	timelineH := handlerPublic.NewTimelineHandler(postSvc)

	app.Get("/", homeH.Handle)
	app.Get("/posts/:slug", postH.Show)
	app.Get("/categories", categoryH.List)
	app.Get("/categories/:slug", categoryH.Show)
	app.Get("/timeline", timelineH.Handle)
	app.Get("/about", handlerPublic.AboutHandler)

	// ─── Studio routes ────────────────────────────────────────────────────────
	authH      := handlerStudio.NewAuthHandler(authSvc, cfg)
	dashboardH := handlerStudio.NewDashboardHandler(postSvc, analyticsSvc)
	postsH     := handlerStudio.NewPostsHandler(postSvc, mediaSvc)
	metricsH   := handlerStudio.NewMetricsHandler(analyticsSvc)

	studio := app.Group("/studio")

	// Auth routes (no auth middleware, but login POST is rate-limited)
	studio.Get("/login", authH.ShowLogin)
	studio.Post("/login", rateLimiter.Middleware(), authH.ProcessLogin)
	studio.Post("/logout", authMW, authH.Logout)

	// Redirect /studio → /studio/dashboard
	studio.Get("/", func(c *fiber.Ctx) error {
		return c.Redirect("/studio/dashboard", fiber.StatusSeeOther)
	})

	// Protected routes
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

	log.Printf("Starting server on :%s (env=%s, csp=%s)", cfg.AppPort, cfg.AppEnv, cfg.CSPMode)
	log.Fatal(app.Listen(":" + cfg.AppPort))
}

func errorHandler(c *fiber.Ctx, err error) error {
	code := fiber.StatusInternalServerError
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}
	if code == fiber.StatusNotFound {
		return c.Status(code).Render("public/404", fiber.Map{
			"Title": "Page not found",
		}, "layouts/base")
	}
	return c.Status(code).SendString("Internal Server Error")
}
