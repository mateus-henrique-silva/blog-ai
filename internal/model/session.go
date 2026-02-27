package model

import "time"

type Session struct {
	ID        string
	UserID    int64
	Data      string // AES-256-GCM encrypted JSON blob
	IPHash    string
	UserAgent string
	ExpiresAt time.Time
	CreatedAt time.Time
}

func (s *Session) IsExpired() bool {
	return s.ExpiresAt.Before(timeNow())
}

// timeNow is a variable so tests can override it.
var timeNow = func() time.Time { return time.Now() }
