package repository

import (
	"database/sql"
	"errors"

	"github.com/mhtecdev/blog-ai/internal/model"
)

var ErrNotFound = errors.New("not found")

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) GetByUsername(username string) (*model.AdminUser, error) {
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, email, created_at, updated_at
		 FROM admin_users WHERE username = ?`, username)

	u := &model.AdminUser{}
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *UserRepo) GetByID(id int64) (*model.AdminUser, error) {
	row := r.db.QueryRow(
		`SELECT id, username, password_hash, email, created_at, updated_at
		 FROM admin_users WHERE id = ?`, id)

	u := &model.AdminUser{}
	err := row.Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Email, &u.CreatedAt, &u.UpdatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	return u, err
}

func (r *UserRepo) Create(username, passwordHash, email string) (*model.AdminUser, error) {
	res, err := r.db.Exec(
		`INSERT INTO admin_users (username, password_hash, email) VALUES (?, ?, ?)`,
		username, passwordHash, email)
	if err != nil {
		return nil, err
	}
	id, _ := res.LastInsertId()
	return r.GetByID(id)
}
