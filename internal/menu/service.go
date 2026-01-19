package menu

import (
	"context"
	"fmt"
	"mime/multipart"

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

	key := fmt.Sprintf(
		"menus/%d/%s",
		restaurantID,
		uuid.New().String(),
	)

	url, err := s.storage.Upload(ctx, key, file)
	if err != nil {
		return 0, "", err
	}

	menuUploadID, err := s.repo.CreateUpload(
		restaurantID,
		url,
		filename,
	)
	if err != nil {
		return 0, "", err
	}

	return menuUploadID, url, nil
}
