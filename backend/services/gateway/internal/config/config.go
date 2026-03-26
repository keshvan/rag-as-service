package config

import (
	"fmt"

	commonCfg "github.com/keshvan/rag-as-service/backend/pkg/common/config"
)

type GatewayConfig struct {
	commonCfg.BaseConfig
	AuthHost  string `env:"AUTH_GRPC_HOST" env-default:"auth-service"`
	AuthPort  int    `env:"AUTH_GRPC_PORT" env-default:"50051"`
	JWTSecret string `env:"JWT_SECRET" env-required:"true"`
}

func MustLoad() *GatewayConfig {
	return commonCfg.MustLoad[GatewayConfig]()
}

func (c *GatewayConfig) AuthAddress() string {
	return fmt.Sprintf("%s:%d", c.AuthHost, c.AuthPort)
}
