package repository

import (
	"database/sql"

	"github.com/mhtecdev/blog-ai/internal/model"
)

type AnalyticsRepo struct {
	db *sql.DB
}

func NewAnalyticsRepo(db *sql.DB) *AnalyticsRepo {
	return &AnalyticsRepo{db: db}
}

func (r *AnalyticsRepo) RecordView(postID int64, ipHash, userAgent string) error {
	_, err := r.db.Exec(
		`INSERT INTO page_views (post_id, ip_hash, user_agent) VALUES (?, ?, ?)`,
		postID, ipHash, userAgent)
	return err
}

func (r *AnalyticsRepo) GetPostMetrics() ([]*model.PostMetric, error) {
	rows, err := r.db.Query(`
		SELECT p.id, p.title, p.slug, COUNT(pv.id) AS view_count
		FROM posts p
		LEFT JOIN page_views pv ON pv.post_id = p.id
		WHERE p.status = 'published'
		GROUP BY p.id
		ORDER BY view_count DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []*model.PostMetric
	for rows.Next() {
		m := &model.PostMetric{}
		if err := rows.Scan(&m.PostID, &m.Title, &m.Slug, &m.ViewCount); err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}
	return metrics, rows.Err()
}

func (r *AnalyticsRepo) GetViewsOverTime(postID int64, days int) ([]*model.DailyCount, error) {
	rows, err := r.db.Query(`
		SELECT date(viewed_at) AS day, COUNT(*) AS count
		FROM page_views
		WHERE post_id = ? AND viewed_at >= date('now', ?)
		GROUP BY day ORDER BY day ASC`,
		postID, "-"+intToStr(days)+" days")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []*model.DailyCount
	for rows.Next() {
		c := &model.DailyCount{}
		if err := rows.Scan(&c.Day, &c.Count); err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	return counts, rows.Err()
}

func (r *AnalyticsRepo) TotalViews() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM page_views`).Scan(&count)
	return count, err
}

func (r *AnalyticsRepo) TotalViewsToday() (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM page_views WHERE date(viewed_at) = date('now')`).Scan(&count)
	return count, err
}

func intToStr(n int) string {
	if n < 0 {
		return "-" + intToStr(-n)
	}
	if n == 0 {
		return "0"
	}
	digits := []byte{}
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func (r *AnalyticsRepo) GetAllPostMetrics() ([]*model.PostMetric, error) {
	return r.GetPostMetrics()
}

func (r *AnalyticsRepo) GetViewsByPost(postID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM page_views WHERE post_id = ?`, postID).Scan(&count)
	return count, err
}

// UniqueViewsByPost counts distinct IPs per post.
func (r *AnalyticsRepo) UniqueViewsByPost(postID int64) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		`SELECT COUNT(DISTINCT ip_hash) FROM page_views WHERE post_id = ?`, postID).Scan(&count)
	return count, err
}

func (r *AnalyticsRepo) GetRecentViews(limit int) ([]*model.DailyCount, error) {
	rows, err := r.db.Query(`
		SELECT date(viewed_at) AS day, COUNT(*) AS count
		FROM page_views
		WHERE viewed_at >= date('now', ?)
		GROUP BY day ORDER BY day ASC`,
		"-"+intToStr(limit)+" days")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []*model.DailyCount
	for rows.Next() {
		c := &model.DailyCount{}
		if err := rows.Scan(&c.Day, &c.Count); err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	return counts, rows.Err()
}

func (r *AnalyticsRepo) TotalPosts() (int64, error) {
	var count int64
	err := r.db.QueryRow(`SELECT COUNT(*) FROM posts WHERE status = 'published'`).Scan(&count)
	return count, err
}
