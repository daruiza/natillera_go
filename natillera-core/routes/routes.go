package routes

import (
	"fmt"
	"time"

	"natillera-core/app/handler"
	"natillera-core/app/repository"
	"natillera-core/config"
	"natillera-core/db"
	natsManager "natillera-shared/nats"
	"natillera-shared/utils"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"github.com/go-chi/cors"
)

func SetupChiRoutes(r *chi.Mux, cnf *config.MicroConfig) {

	// Servicio de Nats
	opts := &natsManager.OptsNats{
		NameStream: cnf.Get("NATS_STREAM"),
		Subjects:   []string{"natillera.login.*", "natillera.upload.*", "natillera.delete.*", "natillera.message.*"},
		MaxAge:     3600 * time.Minute,
	}

	natsService, err := connectNATS(cnf.Get("NATS_URL"), opts)
	if err != nil {
		utils.Error.Printf("error al conectar al servicio de Nats: %v", err)
		return
	}

	utils.Info.Println("NATS service started")
	defer natsService.ManagerDataNats.Close()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	//set route base
	r.Route("/"+cnf.Get("API_ROUTE_BASE"), func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			repository := repository.NewUserRepository(db.GetMongoClient().Database(cnf.Get("MONGO_DATABASE")), "users")
			userHandler := handler.NewUserHandler(repository, natsService, cnf)
			r.Post("/login", userHandler.Login)
			r.Post("/google", userHandler.GoogleLogin)
			r.Get("/google/callback", userHandler.GoogleCallback)
			r.Post("/google/validate", userHandler.ValidateGoogleToken)
		})
	})

}

func connectNATS(natsURL string, opts *natsManager.OptsNats) (*natsManager.NatsStarter, error) {
	nc := natsManager.NewStartNats(natsURL, opts)
	if nc == nil {
		return nil, fmt.Errorf("error al conectar al servicio de Nats con la url: %s", natsURL)
	}
	return nc, nil
}
