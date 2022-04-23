package redis_test

import (
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/skema-dev/skema-go/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/skema-dev/skema-go/config"
	"github.com/stretchr/testify/assert"
)

var mockRedis *miniredis.Miniredis
var conf *config.Config
var redisMgr *redis.RedisManager
var once sync.Once

func createNewRedisClient(key string) *redis.RedisClient {
	// mock server
	once.Do(func() {
		mockRedis, _ = miniredis.Run()
		yaml := "address: " + mockRedis.Addr()
		redisMgr = redis.NewRedisManager()
		conf = config.NewConfigWithString(yaml)
	})

	redisMgr.AddRedisClientsInPool(conf, key)
	redisInstance := redisMgr.GetRedis(key)
	return redisInstance
}

func TestRedis(t *testing.T) {
	redisInstance := createNewRedisClient("test1")
	err := redisInstance.Set("key", "value123", 10)
	assert.Nil(t, err)

	val, err := redisInstance.Get("key")
	assert.Nil(t, err)
	assert.Equal(t, "value123", val)
}

func TestRedisHSetGet(t *testing.T) {
	redisClient := createNewRedisClient("test2")

	redisClient.HSet("key1", map[string]interface{}{
		"name":  "abc",
		"value": 100,
	})

	result, _ := redisClient.HMGet("key1", "name", "value")
	assert.Equal(t, "abc", result[0])
	v, _ := strconv.Atoi(result[1].(string))
	assert.Equal(t, 100, v)

	data, _ := redisClient.HGetAll("key1")
	assert.Equal(t, "abc", data["name"])
	assert.Equal(t, "100", data["value"])
}

func TestPubsub(t *testing.T) {
	redisClient := createNewRedisClient("test2")

	redisClient.Subscribe("ch1_2", "ch1", "ch2")
	redisClient.Subscribe("ch1", "ch1")
	redisClient.Subscribe("ch2", "ch2")

	redisClient.Publish("ch1", "msg11")
	redisClient.Publish("ch2", "msg21")
	redisClient.Publish("ch1", "msg12")

	ch, msg, _ := redisClient.Receive("ch1_2")
	assert.Equal(t, "ch1", ch)
	assert.Equal(t, "msg11", msg)

	ch, msg, _ = redisClient.Receive("ch1")
	assert.Equal(t, "ch1", ch)
	assert.Equal(t, "msg11", msg)

	ch, msg, _ = redisClient.Receive("ch1_2")
	assert.Equal(t, "ch2", ch)
	assert.Equal(t, "msg21", msg)

	ch, msg, _ = redisClient.Receive("ch3")
	assert.Equal(t, "", ch)
	assert.Equal(t, "", msg)

	ch, msg, _ = redisClient.Receive("ch1_2")
	assert.Equal(t, "ch1", ch)
	assert.Equal(t, "msg12", msg)

	timeout1 := make(chan bool, 1)
	timeout2 := make(chan string, 2)

	ch = ""
	msg = ""

	go func() {
		time.Sleep(1 * time.Second)
		timeout1 <- true
	}()
	go func() {
		ch, msg, _ = redisClient.Receive("ch1_2")
		timeout2 <- msg
	}()

	select {
	case <-timeout1:
		assert.Equal(t, "", ch)
		assert.Equal(t, "", msg)
	case v := <-timeout2:
		assert.Fail(t, "message should be empty, but get "+v)
	}
}
