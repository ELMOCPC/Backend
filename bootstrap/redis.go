package bootstrap

import (
	"strconv"

	"github.com/redis/go-redis/v9"
)

func ProvideRedisClient(env *Env) *redis.Client {
	dbNumber, _ := strconv.Atoi(env.Cache.RDBNumber)

	client := redis.NewClient(&redis.Options{
		Addr:     env.Cache.Address + ":" + env.Cache.Port,
		Password: env.Cache.Password,
		DB:       dbNumber,
	})

	return client
}
