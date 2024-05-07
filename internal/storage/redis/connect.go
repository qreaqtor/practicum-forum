package redis

import (
	"github.com/go-redis/redis"
)

func GetConnect(dbName int, password, addr string) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password, // no password set
		DB:       dbName,   // use default DB
	})
	return client
}
