package redis

import (
	"context"
	"github.com/alicebob/miniredis/v2"
	"github.com/skema-dev/skema-go/config"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestRedis(t *testing.T) {
	// mock server
	mockRedis, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mockRedis.Close()

	ctx := context.Background()

	yaml := "address: " + mockRedis.Addr()
	redisMgr := NewRedisManager()
	config := config.NewConfigWithString(yaml)
	redisMgr.AddRedisClientsInPool(config, "db.redis")
	redisInstance := redisMgr.GetRedis("db.redis")

	err = redisInstance.Set(ctx, "key", "value123", 0).Err()
	assert.Nil(t, err)

	val, err := redisInstance.Get(ctx, "key").Result()
	assert.Nil(t, err)
	assert.Equal(t, "value123", val)
}
