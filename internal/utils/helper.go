package utils

import (
	"crypto/rand"
	"encoding/base64"
	"gin/user-management-api/pkg/loggers"
	"log"
	"os"
	"path/filepath"
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

func NewLoggerWithPath(fileName string, level string) *zerolog.Logger {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get working dir", err)
	}
	path := filepath.Join(cwd, "internal/logs", fileName)
	config := loggers.LoggerConfig{
		Level:      level,
		Filename:   path,
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     5,
		Compress:   true,
		IsDev:      GetEnv("APP_ENV", "development"),
	}

	return loggers.NewLogger(config)
}

func GenerateRandomString(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func MustGetWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get working dif", err)
	}
	return dir
}
