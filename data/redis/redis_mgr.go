package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

var (
	redisMgr *RedisManager
)

type RedisManager struct {
	redisPool map[string]*redis.Client
}

func Manager() *RedisManager {
	return redisMgr
}

func NewRedisManager() *RedisManager {
	man := &RedisManager{
		redisPool: map[string]*redis.Client{},
	}
	return man
}

// InitRedisWithConfigFile read redis config and create redis client
func InitRedisWithConfigFile(filepath string, key string) {
	conf := config.NewConfigWithFile(filepath)
	InitRedisWithConfig(conf, key)
}

// InitRedisWithConfig ...
func InitRedisWithConfig(conf *config.Config, key string) {
	redisMgr = NewRedisManager().ReadWithConfig(conf, key)
}

// ReadWithConfig ...
func (d *RedisManager) ReadWithConfig(conf *config.Config, key string) *RedisManager {
	if conf == nil {
		return d
	}

	configs := conf.GetMapConfig(key)
	// get all redis config info, and create redisClients putting in redisPool
	for k, v := range configs {
		d.AddRedisClientsInPool(&v, k)
	}

	return d
}

// AddRedisClientsInPool create redisClients putting in redisPool
func (d *RedisManager) AddRedisClientsInPool(conf *config.Config, key string) {
	if key == "" {
		logging.Fatalf("A redis datasource key must be specified!")
	}

	rdb, err := NewRedis(conf)
	if err != nil {
		logging.Fatalf("failed creating redis client")
	}

	d.redisPool[key] = rdb
}

// GetRedis get a redis client
func (d *RedisManager) GetRedis(key string) *redis.Client {
	if key == "" {
		logging.Fatalf("Key can not be empty!")
	}

	// get redisInstance from redisPoll
	rdb, ok := d.redisPool[key]
	if !ok {
		logging.Fatalf("Get a redis client failed with key: %s", key)
		return nil
	}
	return rdb
}
