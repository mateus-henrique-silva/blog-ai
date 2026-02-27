package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/mhtecdev/blog-ai/internal/model"
)

type PostRepo struct {
	db *sql.DB
}

func NewPostRepo(db *sql.DB) *PostRepo {
	return &PostRepo{db: db}
}

func (r *PostRepo) Create(p *model.Post) (*model.Post, error) {
	res, err := r.db.Exec(
		`INSERT INTO posts (title, slug, excerpt, content_md, content_html, cover_image, category, tags, status, published_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		p.Title, p.Slug, p.Excerpt, p.ContentMD, p.ContentHTML,
		p.CoverImage, p.Category, p.Tags, p.Status, nullTime(p.PublishedAt))
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}

func (r *PostRepo) Update(p *model.Post) (*model.Post, error) {
	_, err := r.db.Exec(
		`UPDATE posts SET title=?, slug=?, excerpt=?, content_md=?, content_html=?,
		 cover_image=?, category=?, tags=?, status=?, published_at=?,
		 updated_at=strftime('%Y-%m-%dT%H:%M:%SZ','now')
		 WHERE id=?`,
		p.Title, p.Slug, p.Excerpt, p.ContentMD, p.ContentHTML,
		p.CoverImage, p.Category, p.Tags, p.Status, nullTime(p.PublishedAt), p.ID)
	if err != nil {
		return nil, err
	}
	return r.GetByID(p.ID)
}

func (r *PostRepo) Delete(id int64) error {
	_, err := r.db.Exec(`DELETE FROM posts WHERE id = ?`, id)
	return err
}

func (r *PostRepo) GetByID(id int64) (*model.Post, error) {
	row := r.db.QueryRow(`SELECT `+postCols+` FROM posts WHERE id = ?`, id)
	return scanPost(row)
}

func (r *PostRepo) GetBySlug(slug string) (*model.Post, error) {
	row := r.db.QueryRow(`SELECT `+postCols+` FROM posts WHERE slug = ?`, slug)
	return scanPost(row)
}

func (r *PostRepo) ListAll() ([]*model.Post, error) {
	rows, err := r.db.Query(
		`SELECT ` + postCols + ` FROM posts ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *PostRepo) ListPublished() ([]*model.Post, error) {
	rows, err := r.db.Query(
		`SELECT ` + postCols + ` FROM posts WHERE status='published' ORDER BY published_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *PostRepo) ListPublishedByCategory(category string) ([]*model.Post, error) {
	rows, err := r.db.Query(
		`SELECT `+postCols+` FROM posts WHERE status='published' AND category=? ORDER BY published_at DESC`,
		category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanPosts(rows)
}

func (r *PostRepo) ListCategories() ([]string, error) {
	rows, err := r.db.Query(
		`SELECT DISTINCT category FROM posts WHERE status='published' AND category != '' ORDER BY category`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cats []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		cats = append(cats, c)
	}
	return cats, rows.Err()
}

func (r *PostRepo) SlugExists(slug string) (bool, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM posts WHERE slug = ?`, slug).Scan(&count)
	return count > 0, err
}

const postCols = `id, uuid, title, slug, excerpt, content_md, content_html,
	cover_image, category, tags, status, published_at, created_at, updated_at`

func scanPost(row *sql.Row) (*model.Post, error) {
	p := &model.Post{}
	var publishedAt, createdAt, updatedAt sql.NullString
	err := row.Scan(
		&p.ID, &p.UUID, &p.Title, &p.Slug, &p.Excerpt,
		&p.ContentMD, &p.ContentHTML, &p.CoverImage,
		&p.Category, &p.Tags, &p.Status,
		&publishedAt, &createdAt, &updatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	if publishedAt.Valid && publishedAt.String != "" {
		t, _ := time.Parse(time.RFC3339, publishedAt.String)
		p.PublishedAt = &t
	}
	p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
	p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
	return p, nil
}

func scanPosts(rows *sql.Rows) ([]*model.Post, error) {
	var posts []*model.Post
	for rows.Next() {
		p := &model.Post{}
		var publishedAt, createdAt, updatedAt sql.NullString
		err := rows.Scan(
			&p.ID, &p.UUID, &p.Title, &p.Slug, &p.Excerpt,
			&p.ContentMD, &p.ContentHTML, &p.CoverImage,
			&p.Category, &p.Tags, &p.Status,
			&publishedAt, &createdAt, &updatedAt)
		if err != nil {
			return nil, err
		}
		if publishedAt.Valid && publishedAt.String != "" {
			t, _ := time.Parse(time.RFC3339, publishedAt.String)
			p.PublishedAt = &t
		}
		p.CreatedAt, _ = time.Parse(time.RFC3339, createdAt.String)
		p.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt.String)
		posts = append(posts, p)
	}
	return posts, rows.Err()
}

func nullTime(t *time.Time) interface{} {
	if t == nil {
		return nil
	}
	return t.UTC().Format(time.RFC3339)
}
