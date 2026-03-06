package db

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DSN = ""
var PostgresClient *gorm.DB

func ConnectPostgresDB(uri string) {
	PostgresClient, err := gorm.Open(postgres.Open(uri), &gorm.Config{})
	if err != nil {
		log.Fatal("Error de conexión a la base de datos")
	}
	log.Println("Base de Datos Postgres Conectada", PostgresClient.Name())
}

func ClosePostgresDB() {
	sqlDB, err := PostgresClient.DB()
	if err != nil {
		log.Fatal("Error al cerrar la conexión a la base de datos")
	}
	err = sqlDB.Close()
	if err != nil {
		log.Fatal("Error al cerrar la conexión a la base de datos")
	}
	log.Println("Base de Datos Desconectada")
}

func AutoMigrate() {
	// Migrar la base de datos SQL GORM
	PostgresClient.AutoMigrate(
	// &models.User{},
	// &models.Acount{},
	// &models.Bank{},
	// &models.Transaction{},
	)
}
