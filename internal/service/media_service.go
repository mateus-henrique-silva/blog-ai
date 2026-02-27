package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/mhtecdev/blog-ai/internal/config"
	"github.com/mhtecdev/blog-ai/internal/model"
	"github.com/mhtecdev/blog-ai/internal/repository"
)

var allowedMIMEs = map[string]string{
	"image/jpeg":  ".jpg",
	"image/png":   ".png",
	"image/gif":   ".gif",
	"image/webp":  ".webp",
	"video/mp4":   ".mp4",
	"video/webm":  ".webm",
	"audio/mpeg":  ".mp3",
	"audio/ogg":   ".ogg",
}

type MediaService struct {
	repo *repository.MediaRepo
	cfg  *config.Config
}

func NewMediaService(repo *repository.MediaRepo, cfg *config.Config) *MediaService {
	return &MediaService{repo: repo, cfg: cfg}
}

func (s *MediaService) Upload(fh *multipart.FileHeader, uploaderID int64) (*model.Media, error) {
	mimeType := fh.Header.Get("Content-Type")
	// Strip parameters (e.g. "image/jpeg; charset=utf-8" → "image/jpeg")
	if idx := strings.Index(mimeType, ";"); idx != -1 {
		mimeType = strings.TrimSpace(mimeType[:idx])
	}

	ext, ok := allowedMIMEs[mimeType]
	if !ok {
		return nil, errors.New("unsupported file type")
	}

	maxBytes := s.cfg.UploadMaxMB * 1024 * 1024
	if fh.Size > maxBytes {
		return nil, fmt.Errorf("file too large (max %dMB)", s.cfg.UploadMaxMB)
	}

	filename := uuid.New().String() + ext
	destPath := filepath.Join(s.cfg.UploadDir, filename)

	if err := os.MkdirAll(s.cfg.UploadDir, 0o755); err != nil {
		return nil, err
	}

	src, err := fh.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	dst, err := os.Create(destPath)
	if err != nil {
		return nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, src); err != nil {
		return nil, err
	}

	media := &model.Media{
		Filename:   filename,
		Original:   fh.Filename,
		MimeType:   mimeType,
		SizeBytes:  fh.Size,
		URL:        "/static/uploads/" + filename,
		UploadedBy: uploaderID,
	}

	return s.repo.Create(media)
}
