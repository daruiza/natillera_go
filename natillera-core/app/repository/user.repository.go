package repository

import (
	"context"
	"fmt"
	"natillera-core/app/domain"
	"natillera-shared/utils"
	"time"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type UserRepository struct {
	database       *mongo.Database
	collection     *mongo.Collection
	CollectionName string
	Validator      *validator.Validate
}

func NewUserRepository(db *mongo.Database, collection string) *UserRepository {
	return &UserRepository{
		database:       db,
		collection:     db.Collection(collection),
		CollectionName: collection,
		Validator:      utils.Validate,
	}
}

func (r *UserRepository) GetUsers(filter *bson.M) (*[]domain.UserResponse, error) {
	if r.collection == nil {
		return nil, fmt.Errorf("collection not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	findOptions := options.Find()

	cursor, err := r.collection.Find(ctx, filter, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find users: %s", err.Error())
	}
	defer cursor.Close(ctx)

	//Retoranamos el UserResponde
	var users []domain.UserResponse
	if err = cursor.All(ctx, &users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	if len(users) == 0 {
		return &users, nil
	}

	//Poblar el rol de cada cuenta
	rolCollection := r.database.Collection("rols")
	//Vamosa iterar users para llenar el rol

	for i := range users {
		rolFilter := bson.D{{Key: "_id", Value: users[i].RolId}}
		var rol domain.Rol
		err := rolCollection.FindOne(ctx, rolFilter).Decode(&rol)
		if err != nil {
			utils.Error.Println("Error to find rol:", err)
			continue // Continúa al siguiente usuario en caso de error al buscar el rol
		}
		// defer rolCursor.Close(ctx)
		users[i].Rol = rol

		// Formateamos las opciones del rol
		if len(users[i].Rol.Options) > 0 {
			var optionsrol []domain.OptionRol
			for _, option := range users[i].Rol.Options {
				optionsrol = append(optionsrol, domain.OptionRol{
					RolId:  users[i].Rol.ID,
					Option: option,
				},
				)
			}
			users[i].Rol.OptionsRol = optionsrol
		}
	}

	return &users, nil
}
