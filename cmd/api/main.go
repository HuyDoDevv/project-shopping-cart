package main

import (
	"gin/user-management-api/internal/app"
	"gin/user-management-api/internal/config"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

func main() {
	LoadEnv()

	// Initialize the configuration
	config := config.NewConfig()

	// Initialize application
	application := app.NewApplication(config)

	// Start server
	if err := application.Run(); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}

func LoadEnv() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal("Unable to get working dif", err)
	}
	envPath := filepath.Join(cwd, ".env")
	err = godotenv.Load(envPath)
	if err != nil {
		log.Println("No .env file found")
	} else {
		log.Println("Load successfully .env")
	}
}
