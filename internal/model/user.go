package model

import "time"

type AdminUser struct {
	ID           int64
	Username     string
	PasswordHash string
	Email        string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
