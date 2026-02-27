package studio

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/model"
	"github.com/mhtecdev/blog-ai/internal/service"
)

type MetricsHandler struct {
	analytics *service.AnalyticsService
}

func NewMetricsHandler(analytics *service.AnalyticsService) *MetricsHandler {
	return &MetricsHandler{analytics: analytics}
}

func (h *MetricsHandler) Handle(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.AdminUser)

	totalViews, _ := h.analytics.TotalViews()
	todayViews, _ := h.analytics.TotalViewsToday()
	totalPosts, _ := h.analytics.TotalPublishedPosts()
	postMetrics, _ := h.analytics.GetPostMetrics()
	recentViews, _ := h.analytics.GetRecentViews(30)

	return c.Render("studio/metrics", fiber.Map{
		"Title":       "Metrics",
		"Section":     "metrics",
		"User":        user,
		"TotalViews":  totalViews,
		"TodayViews":  todayViews,
		"TotalPosts":  totalPosts,
		"PostMetrics": postMetrics,
		"RecentViews": recentViews,
	}, "layouts/studio")
}
