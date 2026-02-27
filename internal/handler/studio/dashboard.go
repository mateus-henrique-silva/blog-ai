package studio

import (
	"github.com/gofiber/fiber/v2"
	"github.com/mhtecdev/blog-ai/internal/model"
	"github.com/mhtecdev/blog-ai/internal/service"
)

type DashboardHandler struct {
	posts     *service.PostService
	analytics *service.AnalyticsService
}

func NewDashboardHandler(posts *service.PostService, analytics *service.AnalyticsService) *DashboardHandler {
	return &DashboardHandler{posts: posts, analytics: analytics}
}

func (h *DashboardHandler) Handle(c *fiber.Ctx) error {
	user := c.Locals("user").(*model.AdminUser)

	totalViews, _ := h.analytics.TotalViews()
	todayViews, _ := h.analytics.TotalViewsToday()
	totalPosts, _ := h.analytics.TotalPublishedPosts()
	topPosts, _ := h.analytics.GetPostMetrics()
	recentViews, _ := h.analytics.GetRecentViews(30)

	if len(topPosts) > 5 {
		topPosts = topPosts[:5]
	}

	return c.Render("studio/dashboard", fiber.Map{
		"Title":       "Dashboard",
		"Section":     "dashboard",
		"User":        user,
		"TotalViews":  totalViews,
		"TodayViews":  todayViews,
		"TotalPosts":  totalPosts,
		"TopPosts":    topPosts,
		"RecentViews": recentViews,
	}, "layouts/studio")
}
