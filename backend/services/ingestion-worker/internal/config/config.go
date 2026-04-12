package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	KafkaBrokers       []string
	KafkaTopic         string
	KafkaGroupID       string
	DatabaseURL        string
	ProcessingSleepMS  int
}

// LoadConfig loads configuration from environment variables
func LoadConfig() *Config {
	brokerStr := os.Getenv("KAFKA_BROKERS")
	if brokerStr == "" {
		brokerStr = "localhost:9092"
	}

	return &Config{
		KafkaBrokers: strings.Split(brokerStr, ","),
		KafkaTopic: getEnvOrDefault("KAFKA_TOPIC", "document.uploaded"),
		KafkaGroupID: getEnvOrDefault("KAFKA_GROUP_ID", "ingestion-worker"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		ProcessingSleepMS: getEnvAsInt("PROCESSING_SLEEP_MS", 200),
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
