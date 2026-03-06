package domain

import "go.mongodb.org/mongo-driver/bson/primitive"

type UserResponse struct {
	ID        primitive.ObjectID `json:"id,omitempty" bson:"_id,omitempty"`
	Email     string             `json:"email" bson:"email" `
	Name      string             `json:"name" bson:"name"`
	Names     string             `json:"names" bson:"names"`
	Lastnames string             `json:"lastnames" bson:"lastnames"`
	Phone     string             `json:"phone" bson:"phone"`
	Theme     string             `json:"theme" bson:"theme"`
	Photo     string             `json:"photo" bson:"photo"`
	Password  string             `json:"password" bson:"password"`
	RolId     primitive.ObjectID `json:"rolId,omitempty" bson:"rol_id,omitempty"`
	CreatedAt primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt primitive.DateTime `json:"updated_at" bson:"updated_at"`
	Rol       Rol                `json:"rol" bson:"rol"`
}

// Estructura para login
type LoginRequest struct {
	// Agregamos 'email' para validar formato correo
	Email string `json:"email" bson:"email" validate:"required,email"`

	// min: longitud mínima, max: longitud máxima
	// containsany: requiere al menos uno de los caracteres especiales listados
	Password string `json:"password" bson:"password" validate:"required,min=8,max=32,containsany=!@#$%^&*"`
}
