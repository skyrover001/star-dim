package utils

import (
	"github.com/go-redis/redis"
)

type Cache struct {
	Redis    *redis.Client
	Address  string
	Password string
	DB       int
}

func (c *Cache) InitRedis() (err error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     c.Address,
		Password: c.Password, // no password set
		DB:       c.DB,       // use default DB
	})

	_, err = rdb.Ping().Result()
	if err != nil {
		return err
	}
	c.Redis = rdb
	return nil
}

func (c *Cache) Set(key string, value interface{}) {
	c.Redis.Set(key, value, -1)
}

func (c *Cache) Get(key string) interface{} {
	return c.Redis.Get(key)
}
