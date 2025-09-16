package config

import (
	"fmt"
	"gin/user-management-api/internal/utils"
	"os"
)

type DatabaseConfig struct {
	Host 				string
	Port 				string
	DBName 			string
	User				string
	Password 		string
	SSLMode 		string
}


type Config struct {
	DB DatabaseConfig
	ServerAddress string
}


func NewConfig() *Config {
	return &Config{
		ServerAddress: fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		DB: DatabaseConfig {
			Host: 			utils.GetEnv("DB_HOST", "localhost"),
			Port: 			utils.GetEnv("DB_PORT", "5433"),
			DBName: 		utils.GetEnv("DB_NAME", "myapp"),
			User: 			utils.GetEnv("DB_USER", "postgres"),
			Password: 	utils.GetEnv("DB_PASSWORD", "postgres"),
			SSLMode: 		utils.GetEnv("DB_SSLMODE", "disable"),
		},
	}
}

func (c *Config) DNS() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}
