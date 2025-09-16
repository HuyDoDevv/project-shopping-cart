package config

import (
	"context"
	"gin/user-management-api/internal/utils"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
		Address string
		Username string
		Password string
		DB int
}


func NewRedisClient() *redis.Client {
	config  := RedisConfig{
		Address: utils.GetEnv("REDIS_ADDRESS","localhost:6379"),
		Username: utils.GetEnv("REDIS_USER",""),
		Password: utils.GetEnv("REDIS_PASSWORD",""),
		DB: utils.GetIntEnv("REDIS_DB",0),
	}
	client := redis.NewClient(&redis.Options{
		Addr: config.Address,
		Username: config.Username,
		Password: config.Password,
		DB: config.DB,
		PoolSize: 20,
		MinIdleConns: 2,
		DialTimeout: 5 * time.Second,
		ReadTimeout: 5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5* time.Second)
	defer cancel()

	_, err := client.Ping(ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis %v", err)
	}

	log.Println("Connected Redis")

	return client
}
