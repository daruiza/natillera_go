package domain

import (
	"natillera-shared/proto/pb_generate_user"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

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
	Rol       Rol                `json:"rol" bson:"rol"`
	CreatedAt primitive.DateTime `json:"created_at" bson:"created_at"`
	UpdatedAt primitive.DateTime `json:"updated_at" bson:"updated_at"`
}

func (receiver UserResponse) ToProtocolBuffer() *pb_generate_user.UserDataPB {
	payload := receiver

	// Creamos las slices para los Options y OptionsRol
	var pbOptions []*pb_generate_user.OptionPB
	for _, opt := range payload.Rol.Options {
		pbOptions = append(pbOptions, &pb_generate_user.OptionPB{
			Name:        opt.Name,
			Description: opt.Description,
		})
	}

	var pbOptionsRol []*pb_generate_user.OptionsRolPB
	for _, opt := range payload.Rol.OptionsRol {
		pbOptionsRol = append(pbOptionsRol, &pb_generate_user.OptionsRolPB{
			RolId: opt.RolId,
			Option: &pb_generate_user.OptionPB{
				Name:        opt.Option.Name,
				Description: opt.Option.Description,
			},
		})
	}

	pbUserData := &pb_generate_user.UserDataPB{
		ID:        payload.ID.Hex(),
		Email:     payload.Email,
		Name:      payload.Name,
		Names:     payload.Names,
		Lastnames: payload.Lastnames,
		Phone:     payload.Phone,
		Theme:     payload.Theme,
		Photo:     payload.Photo,
		Password:  payload.Password,
		RolId:     payload.RolId.Hex(),
		CreatedAt: payload.CreatedAt.Time().GoString(),
		UpdatedAt: payload.UpdatedAt.Time().GoString(),
		Rol: &pb_generate_user.RolPB{
			ID:          payload.Rol.ID,
			Name:        payload.Rol.Name,
			Description: payload.Rol.Description,
			CreatedAt:   payload.Rol.CreatedAt.Time().GoString(),
			UpdatedAt:   payload.Rol.UpdatedAt.Time().GoString(),
			Options:     pbOptions,
			OptionsRol:  pbOptionsRol,
		},
	}
	return pbUserData
}

// Estructura para login
type LoginRequest struct {
	// Agregamos 'email' para validar formato correo
	Email string `json:"email" bson:"email" validate:"required,email"`
	// min: longitud mínima, max: longitud máxima
	// containsany: requiere al menos uno de los caracteres especiales listados
	Password string `json:"password" bson:"password" validate:"required,min=8,max=32,containsany=!@#$%^&*"`
}
