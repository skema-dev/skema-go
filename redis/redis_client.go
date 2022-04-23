package redis

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/skema-dev/skema-go/config"
	"github.com/skema-dev/skema-go/logging"
)

type RedisClient struct {
	redis.Client
	ctx context.Context

	pubsubs map[string]*redis.PubSub
}

// NewRedisClient create a new redis client
func NewRedisClient(config *config.Config) (*RedisClient, error) {
	addr := config.GetString("address")
	password := config.GetString("password")
	db := config.GetInt("db", 0)

	if addr == "" {
		msg := "redis addr cannot be empty"
		logging.Errorf(msg)
		return nil, errors.New(msg)
	}

	client := RedisClient{
		Client: *redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
		ctx:     context.Background(),
		pubsubs: map[string]*redis.PubSub{},
	}

	if ret := client.Ping(context.Background()); ret.Err() != nil {
		return nil, ret.Err()
	}
	logging.Debugf("connecting to redis server %s success!", addr)
	return &client, nil
}

func (r *RedisClient) Get(key string) (string, error) {
	result, err := r.Client.Get(r.ctx, key).Result()
	if err != nil {
		logging.Errorf("Redis Get %s failed: %s", key, err.Error())
	}
	return result, err
}

func (r *RedisClient) Set(key string, value interface{}, expirationInSeconds int64) error {
	err := r.Client.Set(r.ctx, key, value, time.Duration(expirationInSeconds*int64(time.Second))).Err()
	if err != nil {
		logging.Errorf("Redis Set %s failed: %s", key, err.Error())
	}
	return err
}

func (r *RedisClient) HSet(key string, values map[string]interface{}) error {
	err := r.Client.HSet(r.ctx, key, values).Err()
	if err != nil {
		logging.Errorf("Redis HSet %s failed: %s", key, err.Error())
	}
	return err
}

func (r *RedisClient) HGetAll(key string) (map[string]string, error) {
	result, err := r.Client.HGetAll(r.ctx, key).Result()
	if err != nil {
		logging.Errorf("Redis HGet %s failed: %s", key, err.Error())
	}
	return result, err
}

func (r *RedisClient) HMGet(key string, fields ...string) ([]interface{}, error) {
	result, err := r.Client.HMGet(r.ctx, key, fields...).Result()
	if err != nil {
		logging.Errorf("Redis HMGet %s failed: %s", key, err.Error())
	}
	return result, err
}

func (r *RedisClient) Publish(channel string, msg interface{}) error {
	err := r.Client.Publish(r.ctx, channel, msg).Err()
	if err != nil {
		logging.Errorf("Redis Publish to  %s failed: %s", channel, err.Error())
	}
	return err
}

func (r *RedisClient) Subscribe(key string, channels ...string) {
	pubsub := r.Client.Subscribe(r.ctx, channels...)
	r.pubsubs[key] = pubsub
}

func (r *RedisClient) Receive(key string) (string, string, error) {
	pubsub, ok := r.pubsubs[key]
	if !ok {
		err := logging.Errorf("couldn't found pubsub %s", key)
		return "", "", err
	}

	channels := pubsub.Channel()
	for msg := range channels {
		return msg.Channel, msg.Payload, nil
	}

	err := logging.Errorf("no msg received for %s", key)
	return "", "", err
}
