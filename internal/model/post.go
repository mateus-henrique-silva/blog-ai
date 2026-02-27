package model

import "time"

type Post struct {
	ID          int64
	UUID        string
	Title       string
	Slug        string
	Excerpt     string
	ContentMD   string
	ContentHTML string
	CoverImage  string
	Category    string
	Tags        string // comma-separated
	Status      string // "draft" or "published"
	PublishedAt *time.Time
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p *Post) IsPublished() bool {
	return p.Status == "published"
}

func (p *Post) TagList() []string {
	if p.Tags == "" {
		return nil
	}
	var tags []string
	for _, t := range splitCSV(p.Tags) {
		if t != "" {
			tags = append(tags, t)
		}
	}
	return tags
}

func splitCSV(s string) []string {
	var result []string
	start := 0
	for i := 0; i <= len(s); i++ {
		if i == len(s) || s[i] == ',' {
			part := trim(s[start:i])
			if part != "" {
				result = append(result, part)
			}
			start = i + 1
		}
	}
	return result
}

func trim(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t') {
		s = s[:len(s)-1]
	}
	return s
}
