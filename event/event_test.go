package event_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/skema-dev/skema-go/event"
	"github.com/stretchr/testify/assert"
)

func TestEvent(t *testing.T) {
	i := 0
	result := 0
	pubsub := event.NewPubSub()

	end := make(chan int)

	f := func(v interface{}) {
		fmt.Printf("get value %d\n", v.(int))
		i = v.(int)
	}

	e1 := pubsub.Subscribe("test", f)
	pubsub.Subscribe("get", func(interface{}) {
		end <- i
	})

	for k := 0; k < 5; k++ {
		pubsub.Publish("test", k)
	}

	pubsub.Publish("get", 0)
	result = <-end
	assert.Equal(t, 4, result)

	pubsub.Publish("test", 5)
	time.Sleep(1 * time.Second)
	pubsub.Publish("get", 0)
	result = <-end
	assert.Equal(t, 5, result)

	pubsub.Publish("test", 100)
	time.Sleep(1 * time.Second)
	pubsub.Publish("get", 0)
	result = <-end
	assert.Equal(t, 100, result)

	pubsub.Unsubscribe("test", e1)

	pubsub.Publish("test", 200)
	time.Sleep(1 * time.Second)
	pubsub.Publish("get", 0)
	result = <-end
	assert.Equal(t, 100, result)

}
