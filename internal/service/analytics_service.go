package service

import (
	"crypto/sha256"
	"encoding/base64"

	"github.com/mhtecdev/blog-ai/internal/config"
	"github.com/mhtecdev/blog-ai/internal/model"
	"github.com/mhtecdev/blog-ai/internal/repository"
)

type AnalyticsService struct {
	repo *repository.AnalyticsRepo
	cfg  *config.Config
	ch   chan viewEvent
}

type viewEvent struct {
	postID    int64
	ipHash    string
	userAgent string
}

func NewAnalyticsService(repo *repository.AnalyticsRepo, cfg *config.Config) *AnalyticsService {
	svc := &AnalyticsService{
		repo: repo,
		cfg:  cfg,
		ch:   make(chan viewEvent, 256),
	}
	go svc.worker()
	return svc
}

func (s *AnalyticsService) worker() {
	for e := range s.ch {
		_ = s.repo.RecordView(e.postID, e.ipHash, e.userAgent)
	}
}

// RecordView queues a view event asynchronously (non-blocking).
func (s *AnalyticsService) RecordView(postID int64, rawIP, userAgent string) {
	ipHash := hashIPAnalytics(rawIP, s.cfg.IPHashSecret)
	select {
	case s.ch <- viewEvent{postID: postID, ipHash: ipHash, userAgent: userAgent}:
	default:
		// Channel full — drop the event rather than blocking the request
	}
}

func (s *AnalyticsService) GetPostMetrics() ([]*model.PostMetric, error) {
	return s.repo.GetPostMetrics()
}

func (s *AnalyticsService) GetRecentViews(days int) ([]*model.DailyCount, error) {
	return s.repo.GetRecentViews(days)
}

func (s *AnalyticsService) TotalViews() (int64, error) {
	return s.repo.TotalViews()
}

func (s *AnalyticsService) TotalViewsToday() (int64, error) {
	return s.repo.TotalViewsToday()
}

func (s *AnalyticsService) TotalPublishedPosts() (int64, error) {
	return s.repo.TotalPosts()
}

func hashIPAnalytics(ip, secret string) string {
	h := sha256.Sum256([]byte(ip + secret))
	return base64.StdEncoding.EncodeToString(h[:])
}
