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

// --------------------------------------------------
// Upload (unchanged)
// --------------------------------------------------
func (s *Service) UploadMenu(
	ctx context.Context,
	restaurantID int,
	file multipart.File,
	filename string,
) (int, string, error) {

	ext := strings.ToLower(filepath.Ext(filename))

	key := fmt.Sprintf(
		"menus/%d/%s%s",
		restaurantID,
		uuid.New().String(),
		ext,
	)

	if _, err := s.storage.Upload(ctx, key, file); err != nil {
		return 0, "", err
	}

	menuUploadID, err := s.repo.CreateUpload(
		restaurantID,
		key,
		filename,
	)
	if err != nil {
		return 0, "", err
	}

	return menuUploadID, key, nil
}

// --------------------------------------------------
// Parsed menu persistence (NEW, IMPORTANT)
// --------------------------------------------------
func (s *Service) SaveParsedResult(
	menuUploadID int,
	menu *ParsedMenu,
	cost *CostForTwo,
) error {

	doc := map[string]interface{}{
		"items":        menu.Items,
		"tax_percent":  menu.TaxPercent,
		"cost_for_two": cost,
		"version":      "v1",
	}

	return s.repo.SaveParsedMenu(menuUploadID, doc)
}
