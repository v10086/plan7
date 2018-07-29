package cache

import "github.com/go-redis/redis"

var (
	client *redis.Client
)

func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     "192.168.139.128:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
}

func GetCon() *redis.Client {
	return client
}
