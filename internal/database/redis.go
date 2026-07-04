package database

import (
	"context"
	"log"
	"os"

	"github.com/redis/go-redis/v9"
)

// RedisClient is the global instance for the app
var RedisClient *redis.Client
var Ctx = context.Background()

func ConnectRedis() {

	RedisClient = redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASS"),
		DB:       0, // use default DB
	})

	err := RedisClient.Ping(Ctx).Err()
	if err != nil {
		log.Fatalf("Could not connect to Redis Cloud: %v", err)
	}

	log.Println("Connected to Redis Cloud successfully!")
}
