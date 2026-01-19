package menu

type Repository interface {
	Create(menu *Menu) error
	FindByRestaurant(restaurantID string) (*Menu, error)
}
