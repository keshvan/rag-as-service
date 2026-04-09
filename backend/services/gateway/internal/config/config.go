package config

import (
	"fmt"

	commonCfg "github.com/keshvan/rag-as-service/backend/pkg/common/config"
)

type GatewayConfig struct {
	commonCfg.BaseConfig
	AuthHost     string `env:"AUTH_GRPC_HOST" env-default:"auth-service"`
	AuthPort     int    `env:"AUTH_GRPC_PORT" env-default:"50051"`
	DocumentHost string `env:"DOCUMENT_GRPC_HOST" env-default:"document-service"`
	DocumentPort int    `env:"DOCUMENT_GRPC_PORT" env-default:"50051"`
	JWTSecret    string `env:"JWT_SECRET" env-required:"true"`

	S3Endpoint         string `env:"S3_ENDPOINT" env-default:"https://storage.yandexcloud.net"`
	S3Region           string `env:"S3_REGION" env-default:"ru-central1"`
	S3Bucket           string `env:"S3_BUCKET" env-required:"true"`
	S3AccessKeyID      string `env:"S3_ACCESS_KEY_ID" env-required:"true"`
	S3SecretAccessKey  string `env:"S3_SECRET_ACCESS_KEY" env-required:"true"`
	S3PresignTTLSecond int    `env:"S3_PRESIGN_TTL_SECOND" env-default:"900"`
}

func MustLoad() *GatewayConfig {
	return commonCfg.MustLoad[GatewayConfig]()
}

func (c *GatewayConfig) AuthAddress() string {
	return fmt.Sprintf("%s:%d", c.AuthHost, c.AuthPort)
}
