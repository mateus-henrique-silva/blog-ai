package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/mhtecdev/blog-ai/internal/model"
)

type SessionRepo struct {
	db *sql.DB
}

func NewSessionRepo(db *sql.DB) *SessionRepo {
	return &SessionRepo{db: db}
}

func (r *SessionRepo) Create(s *model.Session) error {
	_, err := r.db.Exec(
		`INSERT INTO sessions (id, user_id, data, ip_hash, user_agent, expires_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		s.ID, s.UserID, s.Data, s.IPHash, s.UserAgent, s.ExpiresAt.UTC().Format(time.RFC3339))
	return err
}

func (r *SessionRepo) Get(id string) (*model.Session, error) {
	row := r.db.QueryRow(
		`SELECT id, user_id, data, ip_hash, user_agent, expires_at, created_at
		 FROM sessions WHERE id = ?`, id)

	s := &model.Session{}
	var expiresAt, createdAt string
	err := row.Scan(&s.ID, &s.UserID, &s.Data, &s.IPHash, &s.UserAgent, &expiresAt, &createdAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	s.ExpiresAt, _ = time.Parse(time.RFC3339, expiresAt)
	s.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	return s, nil
}

func (r *SessionRepo) Delete(id string) error {
	_, err := r.db.Exec(`DELETE FROM sessions WHERE id = ?`, id)
	return err
}

func (r *SessionRepo) DeleteExpired() error {
	_, err := r.db.Exec(
		`DELETE FROM sessions WHERE expires_at < ?`,
		time.Now().UTC().Format(time.RFC3339))
	return err
}
