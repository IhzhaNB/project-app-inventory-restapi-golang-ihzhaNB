package sale

// CreateSaleRequest contains data for creating a new sale
type CreateSaleRequest struct {
	Items []SaleItemRequest `json:"items" validate:"required,min=1,dive"`
}

// SaleItemRequest represents a single product in sale
type SaleItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid4"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

// UpdateSaleStatusRequest for changing sale status
type UpdateSaleStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=pending completed cancelled"`
}
