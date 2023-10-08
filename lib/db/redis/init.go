package redis

import (
	"context"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
)

var rdb *redis.Client

type Z = redis.Z

func Init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     viper.GetString("redis.addr"),
		Password: viper.GetString("REDIS_PASSWORD"),
		DB:       viper.GetInt("redis.db"),
	})

	_, err := rdb.Ping(context.TODO()).Result()
	if err != nil {
		panic(err)
	}
}

func Instance() *redis.Client {
	return rdb
}
