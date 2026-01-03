package shelf

import "time"

type ShelfResponse struct {
	ID          string    `json:"id"`
	WarehouseID string    `json:"warehouse_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
