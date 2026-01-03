package warehouse

type CreateWarehouseRequest struct {
	Name    string `json:"name" validate:"required,min=3,max=100"`
	Address string `json:"address" validate:"max=500"`
}

type UpdateWarehouseRequest struct {
	Name    *string `json:"name,omitempty" validate:"omitempty,min=3,max=100"`
	Address *string `json:"address,omitempty" validate:"omitempty,max=500"`
}
