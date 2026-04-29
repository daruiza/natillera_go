module natillera-core

go 1.24.2

require (
	github.com/go-chi/chi v1.5.5
	github.com/go-chi/cors v1.2.2
	github.com/go-playground/validator v9.31.0+incompatible
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	go.mongodb.org/mongo-driver v1.17.9
	golang.org/x/crypto v0.37.0
	golang.org/x/oauth2 v0.35.0
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.31.1
	natillera-shared v0.0.0
)

require (
	cloud.google.com/go/compute/metadata v0.3.0 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.18.0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/nats-io/nats.go v1.48.0 // indirect
	github.com/nats-io/nkeys v0.4.11 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	golang.org/x/sync v0.13.0 // indirect
	golang.org/x/sys v0.32.0 // indirect
	golang.org/x/text v0.24.0 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace natillera-shared => ../natillera-shared
