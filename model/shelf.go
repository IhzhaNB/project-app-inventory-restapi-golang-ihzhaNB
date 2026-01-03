package model

import "github.com/google/uuid"

type Shelf struct {
	BaseModel
	WarehouseID uuid.UUID `db:"warehouse_id" json:"warehouse_id"`
	Code        string    `db:"code" json:"code"`
	Name        string    `db:"name" json:"name"`
}
