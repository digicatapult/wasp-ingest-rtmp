package util

import "os"

const (
	// KafkaBrokersEnv is the kafka brokers environment variable name
	KafkaBrokersEnv = "KAFKA_BROKERS"
	// KafkaTopicEnv is the kafka topic environment variable name
	KafkaTopicEnv = "KAFKA_TOPIC"
)

// GetEnv will lookup a environment variable or return the default
func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
