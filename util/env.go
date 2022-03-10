package util

import "os"

const (
	KafkaBrokersEnv = "KAFKA_BROKERS"
	KafkaTopicEnv   = "KAFKA_TOPIC"
)

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}
