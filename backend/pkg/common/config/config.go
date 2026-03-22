package config

import (
	"fmt"
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTPServerConfig struct {
	Host string `env:"HTTP_HOST" env-default:"0.0.0.0"`
	Port int    `env:"HTTP_PORT" env-default:"8080"`
}

type GRPCServerConfig struct {
	Host string `env:"GRPC_HOST" env-default:"0.0.0.0"`
	Port int    `env:"GRPC_PORT" env-default:"50051"`
}

type PostgresConfig struct {
	Host     string `env:"DB_HOST" env-default:"localhost"`
	Port     int    `env:"DB_PORT" env-default:"5432"`
	User     string `env:"DB_USER" env-default:"postgres"`
	Password string `env:"DB_PASSWORD" env-default:"password"`
	DBName   string `env:"DB_NAME" env-default:"rag_db"`
	SSLMode  string `env:"DB_SSL_MODE" env-default:"disable"`
}

type BaseConfig struct {
	AppEnv   string `env:"APP_ENV" env-default:"development"`
	HTTP     HTTPServerConfig
	GRPC     GRPCServerConfig
	Postgres PostgresConfig
}

func MustLoad[T any]() *T {
	var cfg T

	err := cleanenv.ReadConfig(".env", &cfg)
	if err != nil {
		err = cleanenv.ReadEnv(&cfg)
		if err != nil {
			log.Fatalf("failed to load configuration: %v", err)
		}
	}

	return &cfg
}

func (p PostgresConfig) GetDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		p.Host, p.Port, p.User, p.Password, p.DBName, p.SSLMode)
}

func (h HTTPServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", h.Host, h.Port)
}

func (g GRPCServerConfig) GetAddr() string {
	return fmt.Sprintf("%s:%d", g.Host, g.Port)
}
