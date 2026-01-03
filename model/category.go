package model

type Category struct {
	BaseModel
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description,omitempty"`
}
