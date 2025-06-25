package config

import (
	"os"
	"strconv"
)

// GetEnv retrieves an environment variable or returns a default value if not set
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

// GetIntEnv retrieves an environment variable as an integer.
func GetIntEnv(key string, fallback int) int {
	valStr := GetEnv(key, "")
	val, err := strconv.Atoi(valStr)
	if err != nil || val <= 0 {
		return fallback
	}
	return val
}
