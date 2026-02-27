package model

import "time"

type PageView struct {
	ID        int64
	PostID    int64
	IPHash    string
	UserAgent string
	ViewedAt  time.Time
}

type PostMetric struct {
	PostID    int64
	Title     string
	Slug      string
	ViewCount int64
}

type DailyCount struct {
	Day   string // "2006-01-02"
	Count int64
}
