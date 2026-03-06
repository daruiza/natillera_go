package config

import (
	"errors"
	"natillera-shared/utils"
	"os"
	"strings"

	"github.com/joho/godotenv"
)

type MicroConfig struct {
	data map[string]*string
}

func NewConfig(envs ...string) MicroConfig {
	obj := MicroConfig{
		data: make(map[string]*string),
	}
	obj.Load(envs...)
	return obj
}

func LoadConfig() MicroConfig {

	if err := godotenv.Load(); err != nil {
		utils.Error.Println("no se pudo cargar el archivo .env, usando variables de entorno del sistema")
	}

	return NewConfig(
		"CHI_SERVER_PORT",
		"API_ROUTE_BASE",
		"MONGO_URI",
		"POSTGRES_URI",
		"MONGO_DATABASE",
		"NATS_STREAM",
		"NATS_URL",
		"OAUTH_CLIENT_ID",
		"OAUTH_CLIENT_SECRET",
		"OAUTH_REDIRECT_URL",
	)
}

func (receiver MicroConfig) Get(env string) string {
	data, found := receiver.data[env]
	if !found {
		panic(errors.New("No se encontró la variable en el entorno de configuración " + env))
	}
	return *data
}

func (receiver MicroConfig) Load(envs ...string) {
	for _, env := range envs {
		envValue := receiver.getEnv(env)
		utils.Debug("loading env " + env + ":" + *envValue)
		if envValue != nil {
			receiver.data[env] = envValue
		}
	}
}
func (receiver MicroConfig) getEnv(env string) *string {

	data, found := os.LookupEnv(env)
	if found {
		if len(strings.TrimSpace(data)) == 0 {
			return nil
		}
		return &data
	}
	return nil
}
