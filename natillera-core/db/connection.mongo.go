package db

import (
	"context"
	"natillera-shared/utils"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoClient es la instancia del cliente de MongoDB.
var MongoClient *mongo.Client

// ConnectMongoDB establece la conexión con la base de datos MongoDB.
func ConnectMongoDB(uri string) {
	// URI de conexión a MongoDB. Asegúrate de que coincida con tu configuración.

	// Opciones para la conexión.
	clientOptions := options.Client().ApplyURI(uri)

	// Establecer un timeout para la conexión.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Conectar a MongoDB.
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		utils.Error.Fatal(err)
	}

	// Verificar la conexión haciendo un ping al servidor.
	err = client.Ping(ctx, nil)
	if err != nil {
		utils.Error.Fatal(err)
	}
	utils.Info.Println("Conectado a MongoDB!")
	MongoClient = client
}

// GetMongoClient devuelve la instancia del cliente de MongoDB.
func GetMongoClient() *mongo.Client {
	return MongoClient
}

// CloseMongoDB desconecta el cliente de MongoDB cuando la aplicación termina.
func CloseMongoDB() {
	if MongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := MongoClient.Disconnect(ctx)
		if err != nil {
			utils.Error.Fatal(err)
		}
		utils.Info.Println("Desconectado de MongoDB.")
	}
}
