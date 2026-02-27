package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
	"github.com/mhtecdev/blog-ai/internal/config"
	"github.com/mhtecdev/blog-ai/internal/model"
	"github.com/mhtecdev/blog-ai/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionExpired     = errors.New("session expired")
)

type AuthService struct {
	users    *repository.UserRepo
	sessions *repository.SessionRepo
	cfg      *config.Config
	gcm      cipher.AEAD
}

func NewAuthService(users *repository.UserRepo, sessions *repository.SessionRepo, cfg *config.Config) (*AuthService, error) {
	key := deriveKey(cfg.AppSecret)
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AuthService{users: users, sessions: sessions, cfg: cfg, gcm: gcm}, nil
}

func (s *AuthService) Login(username, password, rawIP, userAgent string) (*model.Session, error) {
	user, err := s.users.GetByUsername(username)
	if errors.Is(err, repository.ErrNotFound) {
		// Constant-time comparison to prevent timing attacks
		_ = bcrypt.CompareHashAndPassword([]byte("$2a$12$fakehash"), []byte(password))
		return nil, ErrInvalidCredentials
	}
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	sessionID := uuid.New().String()
	data, err := s.encrypt(`{"role":"admin"}`)
	if err != nil {
		return nil, err
	}

	session := &model.Session{
		ID:        sessionID,
		UserID:    user.ID,
		Data:      data,
		IPHash:    hashIP(rawIP, s.cfg.IPHashSecret),
		UserAgent: userAgent,
		ExpiresAt: time.Now().Add(s.cfg.SessionDuration),
	}

	if err := s.sessions.Create(session); err != nil {
		return nil, err
	}
	return session, nil
}

func (s *AuthService) Validate(sessionID string) (*model.AdminUser, error) {
	session, err := s.sessions.Get(sessionID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSessionExpired
	}
	if err != nil {
		return nil, err
	}
	if session.IsExpired() {
		_ = s.sessions.Delete(sessionID)
		return nil, ErrSessionExpired
	}
	user, err := s.users.GetByID(session.UserID)
	if errors.Is(err, repository.ErrNotFound) {
		return nil, ErrSessionExpired
	}
	return user, err
}

func (s *AuthService) Logout(sessionID string) error {
	return s.sessions.Delete(sessionID)
}

func (s *AuthService) CreateUser(username, password, email string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return err
	}
	_, err = s.users.Create(username, string(hash), email)
	return err
}

func (s *AuthService) encrypt(plaintext string) (string, error) {
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ct := s.gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ct), nil
}

func (s *AuthService) decrypt(ciphertext string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	nonceSize := s.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}
	nonce, ct := data[:nonceSize], data[nonceSize:]
	pt, err := s.gcm.Open(nil, nonce, ct, nil)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func deriveKey(secret string) []byte {
	h := sha256.Sum256([]byte(secret))
	return h[:]
}

func hashIP(ip, secret string) string {
	h := sha256.Sum256([]byte(ip + secret))
	return base64.StdEncoding.EncodeToString(h[:])
}
