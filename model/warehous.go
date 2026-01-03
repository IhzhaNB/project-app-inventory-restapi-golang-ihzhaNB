package model

type Warehouse struct {
	BaseModel
	Name    string `db:"name" json:"name"`
	Address string `db:"address" json:"address"`
}
