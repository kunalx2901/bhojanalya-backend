package menu

import "context"

// Repository defines all database operations for menus
type Repository interface {

	// -------------------------------
	// Upload & Parsing (SAFE)
	// -------------------------------

	// Create OR replace menu upload for a restaurant
	UpsertUpload(
		ctx context.Context,
		restaurantID int,
		objectKey string,
		filename string,
	) (menuID int, status string, err error)

	// Atomically mark menu as PARSED and save JSON
	MarkParsed(
		ctx context.Context,
		restaurantID int,
		doc map[string]interface{},
	) error

	// Mark menu as FAILED (no parsed_data written)
	MarkFailed(
		ctx context.Context,
		restaurantID int,
		reason string,
	) error

	// Context for competition snapshot
	GetMenuContext(
		ctx context.Context,
		restaurantID int,
	) (city string, cuisine string, err error)

	// -------------------------------
	// Admin Approval
	// -------------------------------

	ListPending(ctx context.Context) ([]MenuUpload, error)
	Approve(ctx context.Context, restaurantID int, adminID string) error
	Reject(ctx context.Context, restaurantID int, adminID string, reason string) error
}

