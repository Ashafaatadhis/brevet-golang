package config

import (
	"context"
	"log"
	"strconv"

	"github.com/redis/go-redis/v9"
)

var (
	// RedisClient is the global Redis client
	RedisClient *redis.Client
	// Ctx is the context for Redis operations
	Ctx = context.Background()
)

// InitRedis initializes the Redis client using environment variables.
func InitRedis() {

	redisAddr := GetEnv("REDIS_ADDR", "localhost:6379")
	redisPassword := GetEnv("REDIS_PASSWORD", "")
	redisDBStr := GetEnv("REDIS_DB", "0")

	redisDB, err := strconv.Atoi(redisDBStr)
	if err != nil {
		redisDB = 0
	}

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisAddr,
		Password: redisPassword,
		DB:       redisDB,
	})

	_, err = RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("failed to connect to Redis: %v", err)
	}
	log.Println("Connected to Redis")
}
