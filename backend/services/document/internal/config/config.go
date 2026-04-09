package config

import commonCfg "github.com/keshvan/rag-as-service/backend/pkg/common/config"

type DocumentConfig struct {
	commonCfg.BaseConfig

	S3Endpoint         string `env:"S3_ENDPOINT" env-default:"https://storage.yandexcloud.net"`
	S3Region           string `env:"S3_REGION" env-default:"ru-central1"`
	S3Bucket           string `env:"S3_BUCKET" env-required:"true"`
	S3AccessKeyID      string `env:"S3_ACCESS_KEY_ID" env-required:"true"`
	S3SecretAccessKey  string `env:"S3_SECRET_ACCESS_KEY" env-required:"true"`
	S3PresignTTLSecond int    `env:"S3_PRESIGN_TTL_SECOND" env-default:"900"`
}

func MustLoad() *DocumentConfig {
	return commonCfg.MustLoad[DocumentConfig]()
}
