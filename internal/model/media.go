package model

import "time"

type Media struct {
	ID         int64
	Filename   string
	Original   string
	MimeType   string
	SizeBytes  int64
	URL        string
	UploadedBy int64
	CreatedAt  time.Time
}
