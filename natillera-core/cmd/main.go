package main

import (
	"natillera-core/config"
	"natillera-core/db"
	"natillera-core/routes"
	"natillera-shared/utils"
	"net/http"

	"github.com/go-chi/chi"
)

func main() {
	//Environments
	configData := config.LoadConfig()

	//Mongo
	db.ConnectMongoDB(configData.Get("MONGO_URI"))
	defer db.CloseMongoDB()

	//Postgres
	// db.ConnectPostgresDB(configData.Get("POSTGRES_URI"))
	// defer db.ClosePostgresDB()

	//Validación
	//Inicializar el validador
	utils.NewValidator()

	r := chi.NewRouter()
	routes.SetupChiRoutes(r, &configData)
	utils.Info.Printf("Server running on port %s", configData.Get("CHI_SERVER_PORT"))
	http.ListenAndServe(":"+configData.Get("CHI_SERVER_PORT"), r)

}
