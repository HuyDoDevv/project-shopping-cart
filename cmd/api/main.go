package main

import (
	"gin/user-management-api/internal/app"
	"gin/user-management-api/internal/config"
)

func main()  {
	// Initialize the configuration
	config := config.NewConfig()

	// Initialize application
	application := app.NewApplication(config)

	// Start server
	if err := application.Run(); err != nil {
		panic("Failed to start server: " + err.Error())
	}
}
