package redis

import (
	"cc.tim/client/config"
	"context"
	"github.com/go-redis/redis/v8"
	"log"
)

var (
	RedisClient *redis.Client
	Ctx         = context.Background()
)

// InitRedis 初始化Redis连接
func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     config.Config.Redis.Addr,
		Password: config.Config.Redis.Password,
		DB:       config.Config.Redis.DB,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}
}
