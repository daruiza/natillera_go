package repository

import (
	"natillera-core/app/domain"
	"natillera-shared/utils"

	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
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
	var users []domain.UserResponse

	return &users, nil
}
