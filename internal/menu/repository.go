package menu

import "context"

// Repository defines all database operations for menus
type Repository interface {

	// -------------------------------
	// Upload & Parsing
	// -------------------------------

	// Create a menu upload entry (raw menu file)
	CreateUpload(
		restaurantID int,
		objectKey string, // R2 object key (NOT public URL)
		filename string,
	) (int, error)

	// Save parsed menu + cost-for-two as JSON
	SaveParsedMenu(
		menuUploadID int,
		doc map[string]interface{},
	) error

	// Get city and cuisine from menu upload entry
	GetMenuContext(
		ctx context.Context,
		menuUploadID int,
	) (city string, cuisine string, err error)

	// -------------------------------
	// Admin Approval (FINAL PHASE)
	// -------------------------------

	// List menus that are parsed but not yet approved
	ListPending(
		ctx context.Context,
	) ([]MenuUpload, error)

	// Approve a parsed menu
	Approve(
		ctx context.Context,
		menuID int,
		adminID string,
	) error

	// Reject a parsed menu with reason
	Reject(
		ctx context.Context,
		menuID int,
		adminID string,
		reason string,
	) error
}
