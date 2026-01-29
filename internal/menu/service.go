package menu

import (
	"context"
	"errors"
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
// Upload Menu (ONE MENU PER RESTAURANT)
// --------------------------------------------------
func (s *Service) UploadMenu(
	ctx context.Context,
	restaurantID int,
	file multipart.File,
	filename string,
) (string, error) {

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "", errors.New("invalid file")
	}

	key := fmt.Sprintf(
		"menus/%d/%s%s",
		restaurantID,
		uuid.New().String(),
		ext,
	)

	if _, err := s.storage.Upload(ctx, key, file); err != nil {
		return "", err
	}

	_, status, err := s.repo.UpsertUpload(
		ctx,
		restaurantID,
		key,
		filename,
	)
	if err != nil {
		return "", err
	}

	if status == "PARSED" {
		return "", errors.New("menu already parsed and locked")
	}

	return key, nil
}

// --------------------------------------------------
// Persist Parsed Menu (ATOMIC)
// --------------------------------------------------
func (s *Service) SaveParsedResult(
	ctx context.Context,
	restaurantID int,
	menu *ParsedMenu,
	cost *CostForTwo,
) error {

	if menu == nil || cost == nil {
		return errors.New("invalid parsed menu data")
	}

	doc := map[string]interface{}{
		"items":        menu.Items,
		"tax_percent":  menu.TaxPercent,
		"cost_for_two": cost,
		"version":      "v1",
	}

	return s.repo.MarkParsed(ctx, restaurantID, doc)
}

// --------------------------------------------------
// Mark Parsing Failure (SAFE RETRY)
// --------------------------------------------------
func (s *Service) MarkParsingFailed(
	ctx context.Context,
	restaurantID int,
	reason string,
) error {
	return s.repo.MarkFailed(ctx, restaurantID, reason)
}

// --------------------------------------------------
// Fetch Menu Context (city + cuisine)
// --------------------------------------------------
func (s *Service) GetMenuContext(
	ctx context.Context,
	restaurantID int,
) (string, string, error) {
	return s.repo.GetMenuContext(ctx, restaurantID)
}

// --------------------------------------------------
// ADMIN APPROVAL — FINAL PHASE
// --------------------------------------------------

// Get menus pending admin approval
func (s *Service) GetPendingMenus(
	ctx context.Context,
) ([]MenuUpload, error) {
	return s.repo.ListPending(ctx)
}

// Approve a parsed menu (ADMIN)
func (s *Service) ApproveMenu(
	ctx context.Context,
	restaurantID int,
	adminID string,
) error {
	return s.repo.Approve(ctx, restaurantID, adminID)
}

// Reject a parsed menu (ADMIN)
func (s *Service) RejectMenu(
	ctx context.Context,
	restaurantID int,
	adminID string,
	reason string,
) error {
	return s.repo.Reject(ctx, restaurantID, adminID, reason)
}


func (s *Service) ApproveRestaurant(
	ctx context.Context,
	restaurantID int,
	adminID string,
) error {
	return s.repo.ApproveByRestaurant(ctx, restaurantID, adminID)
}
