package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	KafkaBrokers      []string
	KafkaTopic        string
	KafkaGroupID      string
	DatabaseURL       string
	ProcessingSleepMS int

	S3Endpoint        string
	S3Region          string
	S3Bucket          string
	S3AccessKeyID     string
	S3SecretAccessKey string

	YandexApiKey   string
	YandexFolderID string
	YandexBaseURL  string

	QdrantHost       string
	QdrantPort       int
	QdrantCollection string
	QdrantUseTLS     bool

	WorkerConcurrency int

	DownloadDir string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	brokerStr := os.Getenv("KAFKA_BROKERS")
	if brokerStr == "" {
		brokerStr = "localhost:9092"
	}

	return &Config{
		KafkaBrokers:      strings.Split(brokerStr, ","),
		KafkaTopic:        getEnvOrDefault("KAFKA_TOPIC", "document.uploaded"),
		KafkaGroupID:      getEnvOrDefault("KAFKA_GROUP_ID", "ingestion-worker"),
		DatabaseURL:       os.Getenv("DATABASE_URL"),
		ProcessingSleepMS: getEnvAsInt("PROCESSING_SLEEP_MS", 200),

		S3Endpoint:        getEnvOrDefault("S3_ENDPOINT", "https://storage.yandexcloud.net"),
		S3Region:          getEnvOrDefault("S3_REGION", "ru-central1"),
		S3Bucket:          os.Getenv("S3_BUCKET"),
		S3AccessKeyID:     os.Getenv("S3_ACCESS_KEY_ID"),
		S3SecretAccessKey: os.Getenv("S3_SECRET_ACCESS_KEY"),

		YandexApiKey:   os.Getenv("YANDEX_API_KEY"),
		YandexFolderID: os.Getenv("YANDEX_FOLDER_ID"),
		YandexBaseURL:  os.Getenv("YANDEX_BASE_URL"),

		QdrantHost:       getEnvOrDefault("QDRANT_HOST", "localhost"),
		QdrantPort:       getEnvAsInt("QDRANT_PORT", 6334),
		QdrantCollection: getEnvOrDefault("QDRANT_COLLECTION", "documents"),
		QdrantUseTLS:     getEnvOrDefault("QDRANT_USE_TLS", "false") == "true",

		WorkerConcurrency: getEnvAsInt("WORKER_CONCURRENCY", 4),

		DownloadDir: getEnvOrDefault("DOWNLOAD_DIR", "/tmp/raas-ingestion"),
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	valStr := os.Getenv(key)
	if valStr == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(valStr)
	if err != nil {
		return defaultValue
	}
	return val
}
