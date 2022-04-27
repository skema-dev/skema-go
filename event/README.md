# Simple Event PubSub Support

Event driven is quite important in some cases. Unfortunately, if you google around, there is no recommended "Event" solution for go project (There is one in ethereum project, but you don't want to clone the whole ethereum codebase just for event package...).  
One of the reasons is because go channel is fundamentally simliar to "event", which you can send and receive message in different gorountines. So it's not very difficult to implement even-alike pub/sub support.   

However, it's still taking time to do so and pretty error prone. Then I made this simple implementing as part of skema-go framework for you to use. In skema-go itself, the event package is use in `data/dao` for notifying elasticsearch to update indexes.  

The usage is simple, you can check `event_test.go` for details. Here is how you subscribe to an event and notify through publish:  
```
	i := 0
	pubsub := event.NewPubSub()

	end := make(chan int)
	f := func(v interface{}) {
		fmt.Printf("get value %d\n", v.(int))
		i = v.(int)
	}

    e1 := pubsub.Subscribe("test", f)	// create subscription (a.k.a. listener)
    pubsub.Subscribe("get", func(interface{}) {
		end <- i
	})

	for k := 0; k < 5; k++ {
		pubsub.Publish("test", k)      // Publish data (send to the listener)
	}

	pubsub.Publish("get", 0)           // notify another subscriber to retrieve value (send to chan variable end)
	result = <-end                     // get the value
	assert.Equal(t, 4, result)

	pubsub.Unsubscribe("test", e1)

```