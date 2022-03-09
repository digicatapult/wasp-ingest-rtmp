package util

import "os"

type UtilOperations interface {
	GetEnv()
}

func GetEnv(key, defaultValue string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}

	return defaultValue
}