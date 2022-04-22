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
func (d *RedisManager) AddRedisClientsInPool(conf *config.Config, dbKey string) {
	if dbKey == "" {
		logging.Fatalf("A redis datasource key must be specified!")
	}

	rdb, err := NewRedis(conf)
	if err != nil {
		logging.Fatalf("failed creating redis client")
	}

	d.redisPool[dbKey] = rdb
}

// GetRedis get a redis client
func (d *RedisManager) GetRedis(dbKey string) *redis.Client {
	// use default as dbKey if dbKey is empty
	if dbKey == "" {
		dbKey = "default"
	}

	// get redisInstance from redisPoll
	rdb, ok := d.redisPool[dbKey]
	if !ok {
		logging.Fatalf("Get a redis client failed!")
		return nil
	}
	return rdb
}
