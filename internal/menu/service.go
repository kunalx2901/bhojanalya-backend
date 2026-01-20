package menu

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

type Storage interface {
	Upload(ctx context.Context, key string, file multipart.File) (string, error)
}

type Service struct {
	repo    Repository
	storage Storage
}

func NewService(repo Repository, storage Storage) *Service {
	return &Service{repo: repo, storage: storage}
}

func (s *Service) UploadMenu(
	ctx context.Context,
	restaurantID int,
	file multipart.File,
	filename string,
) (int, string, error) {
	// Extract file extension
	ext := strings.ToLower(filepath.Ext(filename))

	// Generate unique key (R2 object key)
	key := fmt.Sprintf(
		"menus/%d/%s%s",
		restaurantID,
		uuid.New().String(),
		ext,
	)

	// Upload to R2 (returns public URL, not key)
	_, err := s.storage.Upload(ctx, key, file)
	if err != nil {
		return 0, "", err
	}

	// Store object KEY in database, not the public URL
	menuUploadID, err := s.repo.CreateUpload(
		restaurantID,
		key,           // Store object key
		filename,      // Store original filename
	)
	if err != nil {
		return 0, "", err
	}

	return menuUploadID, key, nil
}
