package presence

import (
	"context"
	"log"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/yourusername/harbor/internal/config"
)

func NewRedisClient(cfg *config.Config) (*redis.Client, error) {
	opt, err := redis.ParseURL(cfg.RedisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	log.Println("Redis connection established")
	return client, nil
}

func CloseRedisClient(client *redis.Client) {
	if client != nil {
		client.Close()
		log.Println("Redis connection closed")
	}
}