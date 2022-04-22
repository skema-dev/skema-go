package redis

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/skema-dev/skema-go/config"
)

// NewRedis create a new redis client
func NewRedis(config *config.Config) (*redis.Client, error) {
	addr := config.GetString("address")
	password := config.GetString("password")
	db := config.GetInt("db")

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if ret := rdb.Ping(context.Background()); ret.Err() != nil {
		return nil, ret.Err()
	}
	return rdb, nil
}
