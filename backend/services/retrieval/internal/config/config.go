package config

import commonCfg "github.com/keshvan/rag-as-service/backend/pkg/common/config"

type RetrievalConfig struct {
	commonCfg.BaseConfig

	YandexAPIKey                string `env:"YANDEX_API_KEY" env-required:"true"`
	YandexFolderID              string `env:"YANDEX_FOLDER_ID" env-required:"true"`
	YandexEmbeddingModel        string `env:"YANDEX_EMBEDDING_MODEL"`
	YandexEmbeddingBaseURL      string `env:"YANDEX_EMBEDDING_BASE_URL" env-default:"https://ai.api.cloud.yandex.net/v1"`
	YandexEmbeddingMaxBatchSize int    `env:"YANDEX_EMBEDDING_MAX_BATCH_SIZE" env-default:"128"`

	QdrantURL            string `env:"QDRANT_URL" env-default:"http://qdrant:6333"`
	QdrantAPIKey         string `env:"QDRANT_API_KEY"`
	QdrantCollection     string `env:"QDRANT_COLLECTION" env-default:"document_chunks"`
	QdrantVectorName     string `env:"QDRANT_VECTOR_NAME"`
	QdrantTimeoutSeconds int    `env:"QDRANT_TIMEOUT_SECONDS" env-default:"10"`

	DefaultLimit int `env:"RETRIEVAL_DEFAULT_LIMIT" env-default:"5"`
	MaxLimit     int `env:"RETRIEVAL_MAX_LIMIT" env-default:"20"`
}

func MustLoad() *RetrievalConfig {
	return commonCfg.MustLoad[RetrievalConfig]()
}
