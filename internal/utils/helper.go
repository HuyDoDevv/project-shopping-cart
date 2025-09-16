package utils

import (
	"gin/user-management-api/pkg/loggers"
	"os"
	"strconv"

	"github.com/rs/zerolog"
)


func GetEnv(key, defaulValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaulValue
}

func GetIntEnv(key string, defaulValue int) int {
	value := os.Getenv(key)
	if value != "" {
		return defaulValue
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return defaulValue
	}

	return intVal
}


func NewLoggerWithPath(path string, level string) *zerolog.Logger {
	config := loggers.LoggerConfig {
		Level: 			level,
		Filename:   path,
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     5,
		Compress:   true,
		IsDev: GetEnv("APP_ENV", "development"),
	}

	return loggers.NewLogger(config)
}
