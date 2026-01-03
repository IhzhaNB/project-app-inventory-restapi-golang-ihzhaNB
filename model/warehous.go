package model

type Warehouse struct {
	BaseModel
	Code    string `db:"code" json:"code"`
	Name    string `db:"name" json:"name"`
	Address string `db:"address" json:"address"`
}
