package kafka_utils

import (
	"os"
	"sync"
)

type KafkaConfig struct {
	KafkaHost string
	KafkaPort string
	Topics    []string
}

var (
	instance *KafkaConfig
	once     sync.Once
)

func GetKafkaConfig() *KafkaConfig {
	once.Do(func() {
		instance = loadConfig()
	})
	return instance
}

func loadConfig() *KafkaConfig {
	// Load kafka_utils host and port
	kafkaHost := os.Getenv("KAFKA_HOST")
	if kafkaHost == "" {
		kafkaHost = "localhost"
	}

	kafkaPort := os.Getenv("KAFKA_PORT")
	if kafkaPort == "" {
		kafkaPort = "9092"
	}

	// Load kafka_utils Topics from environment variables or use default values
	topics := []string{
		loadKafkaTopic(UserRegistrationEventTopic),
		loadKafkaTopic(AccountDeletionEventTopic),
		loadKafkaTopic(AccountDeletionEmailEventTopic),
		loadKafkaTopic(PasswordChangeEmailEventTopic),
		loadKafkaTopic(LoginOTTEmailEventTopic),
	}

	return &KafkaConfig{
		KafkaHost: kafkaHost,
		KafkaPort: kafkaPort,
		Topics:    topics,
	}
}

// loadKafkaTopic returns the environment variable value or the default value
func loadKafkaTopic(topic KafkaTopic) string {
	if value := os.Getenv(topic.envVarName); value != "" {
		topic.DefaultValue = value
	}

	return topic.DefaultValue
}
