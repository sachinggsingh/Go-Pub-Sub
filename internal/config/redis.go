package config

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var ctx = context.Background()

func InitRedis() error {
	addr := os.Getenv("REDIS_ADDR")
	username := os.Getenv("REDIS_USERNAME")
	password := os.Getenv("REDIS_PASSWORD")
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Username: username,
		Password: password,
		DB:       0,
	})

	RDB = client

	if err := client.Set(ctx, "isConnected", "Yes Connected", 0).Err(); err != nil {
		fmt.Println("Not connected to redis")
		return err
	}
	result, err := RDB.Get(ctx, "isConnected").Result()
	if err != nil {
		fmt.Println("Not connected")
		return err
	}
	fmt.Println(result)
	return nil

}
