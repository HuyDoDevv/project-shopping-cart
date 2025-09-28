package config

import (
	"fmt"
	"gin/user-management-api/internal/utils"
	"os"
)

type DatabaseConfig struct {
	Host     string
	Port     string
	DBName   string
	User     string
	Password string
	SSLMode  string
}

type Config struct {
	DB                 DatabaseConfig
	ServerAddress      string
	MailProviderType   string
	MailProviderConfig map[string]any
}

func NewConfig() *Config {
	mailProviderConfig := make(map[string]any)

	mailProviderType := utils.GetEnv("MAIL_PROVIDER_TYPE", "mailtrap")
	if mailProviderType == "mailtrap" {
		mailtrapConfig := map[string]any{
			"mail_sender":      utils.GetEnv("MAILTRAP_SENDER", "huydo@leangolangwithhuy.com"),
			"name_sender":      utils.GetEnv("MAILTRAP_NAME_SENDER", "Support team Code With HuyDo"),
			"mailtrap_url":     utils.GetEnv("MAILTRAP_URL", "https://sandbox.api.mailtrap.io/api/send/4064027"),
			"mailtrap_api_key": utils.GetEnv("MAILTRAP_API_KEY", "0758e60595e2a189ef6bd6bcf4d27e4d"),
		}
		mailProviderConfig["mailtrap"] = mailtrapConfig
	}

	return &Config{
		ServerAddress: fmt.Sprintf(":%s", os.Getenv("SERVER_PORT")),
		DB: DatabaseConfig{
			Host:     utils.GetEnv("DB_HOST", "localhost"),
			Port:     utils.GetEnv("DB_PORT", "5433"),
			DBName:   utils.GetEnv("DB_NAME", "myapp"),
			User:     utils.GetEnv("DB_USER", "postgres"),
			Password: utils.GetEnv("DB_PASSWORD", "postgres"),
			SSLMode:  utils.GetEnv("DB_SSLMODE", "disable"),
		},
		MailProviderType:   mailProviderType,
		MailProviderConfig: mailProviderConfig,
	}
}

func (c *Config) DNS() string {
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", c.DB.Host, c.DB.Port, c.DB.User, c.DB.Password, c.DB.DBName, c.DB.SSLMode)
}
