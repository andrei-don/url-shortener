package config

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

var redisClient = redis.NewClient

func ConnectRedis(addr string) (*redis.Client, error) {
	client := redisClient(&redis.Options{
		Addr:     addr,
		Password: "",
		DB:       0,
	})

	_, err := client.Ping(Ctx).Result()
	if err != nil {
		return nil, fmt.Errorf("redis connection failed: %w", err)
	}

	fmt.Println("Connected to Redis!")
	return client, nil
}
