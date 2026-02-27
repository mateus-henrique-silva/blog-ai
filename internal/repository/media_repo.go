package repository

import (
	"database/sql"
	"errors"

	"github.com/mhtecdev/blog-ai/internal/model"
)

type MediaRepo struct {
	db *sql.DB
}

func NewMediaRepo(db *sql.DB) *MediaRepo {
	return &MediaRepo{db: db}
}

func (r *MediaRepo) Create(m *model.Media) (*model.Media, error) {
	res, err := r.db.Exec(
		`INSERT INTO media (filename, original, mime_type, size_bytes, url, uploaded_by)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		m.Filename, m.Original, m.MimeType, m.SizeBytes, m.URL, m.UploadedBy)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

func (r *MediaRepo) GetByID(id int64) (*model.Media, error) {
	row := r.db.QueryRow(
		`SELECT id, filename, original, mime_type, size_bytes, url, uploaded_by, created_at
		 FROM media WHERE id = ?`, id)
	m := &model.Media{}
	err := row.Scan(&m.ID, &m.Filename, &m.Original, &m.MimeType,
		&m.SizeBytes, &m.URL, &m.UploadedBy, &m.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return m, err
}
