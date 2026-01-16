package restaurant

type Repository interface {
	Create(restaurant *Restaurant) error
	ListByOwner(ownerID string) ([]*Restaurant, error)
}
