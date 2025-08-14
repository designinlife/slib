package os

import (
	"os"
)

func GetEnvDefault[T any](key string, defaultValue T) T {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return To[T](value)
}

func IsEnvEmpty(key string) bool {
	return os.Getenv(key) == ""
}

func To[T any](v any) T {
	val, ok := v.(T)
	if !ok {
		var zero T
		return zero
	}
	return val
}
