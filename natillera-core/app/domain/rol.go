package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type Rol struct {
	ID          string             `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt   primitive.DateTime `json:"updated_at" bson:"updated_at"`
	Options     []Option           `json:"options" bson:"options"`
	OptionsRol  []OptionRol        `json:"optionrols" bson:"optionrols"`
}

type Option struct {
	Name        string `json:"name" bson:"name"`
	Description string `json:"description" bson:"description"`
}

type OptionRol struct {
	RolId  string `json:"rolId" bson:"_id,omitempty"`
	Option Option `json:"option" bson:"option"`
}

type RolResponse struct {
	ID          string             `json:"id" bson:"_id,omitempty"`
	Name        string             `json:"name" bson:"name"`
	Description string             `json:"description" bson:"description"`
	CreatedAt   primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt   primitive.DateTime `json:"updated_at" bson:"updated_at"`
	Options     []Option           `json:"options" bson:"options"`
}
