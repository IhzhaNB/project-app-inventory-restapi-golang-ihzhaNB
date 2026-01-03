package shelf

type CreateShelfRequest struct {
	WarehouseID string `json:"warehouse_id" validate:"required,uuid4"`
	Name        string `json:"name" validate:"required,min=3,max=100"`
}

type UpdateShelfRequest struct {
	WarehouseID *string `json:"warehouse_id,omitempty" validate:"required,uuid4"`
	Name        *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
}
