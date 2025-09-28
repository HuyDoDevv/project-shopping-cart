package main

import (
	"gin/user-management-api/internal/app"
	"gin/user-management-api/internal/config"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/loggers"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	rootDir := mustGetWorkingDir()

	logFile := filepath.Join(rootDir, "internal/logs/app.log")

	loggers.InitLogger(loggers.LoggerConfig{
		Level:      "info",
		Filename:   logFile,
		MaxSize:    1,
		MaxBackups: 5,
		MaxAge:     5,
		Compress:   true,
		IsDev:      utils.GetEnv("APP_ENV", "development"),
	})
	LoadEnv(filepath.Join(rootDir, ".env"))

	// Initialize the configuration
	config := config.NewConfig()

	// Initialize application
	application := app.NewApplication(config)

	// Start server
	if err := application.Run(); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}

func mustGetWorkingDir() string {
	dir, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get working dif", err)
	}
	return dir
}

func LoadEnv(path string) {
	if err := godotenv.Load(path); err != nil {
		loggers.Log.Warn().Msg("No .env file found")
	} else {
		loggers.Log.Info().Msg("Load successfully .env")
	}
}
