package main

import (
	"gin/user-management-api/internal/config"
	"gin/user-management-api/internal/utils"
	"gin/user-management-api/pkg/loggers"
	"path/filepath"

	"github.com/joho/godotenv"
)

func newWorker(cfb *config.Config) {

}

func main() {
	rootDir := utils.MustGetWorkingDir()

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

	if err := godotenv.Load(filepath.Join(rootDir, ".env")); err != nil {
		loggers.Log.Warn().Msg("No .env file found")
	} else {
		loggers.Log.Info().Msg("Load successfully .env in worker")
	}

	// Initialize the configuration
	config := config.NewConfig()

	newWorker(config)
}
