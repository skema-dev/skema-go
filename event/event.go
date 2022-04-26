package event

type EventHandler func(interface{})

type PubSub struct {
	events map[string][]chan interface{}
	stops  map[chan interface{}]bool
}

func NewPubSub() *PubSub {
	pubsub := &PubSub{
		events: map[string][]chan interface{}{},
		stops:  map[chan interface{}]bool{},
	}

	return pubsub
}

// Subscribe to an event with a function handler
// return the channel in case you need to unsubscribe leater
func (p *PubSub) Subscribe(eventName string, h EventHandler) chan interface{} {
	ch := make(chan interface{})

	channels, ok := p.events[eventName]
	if !ok {
		p.events[eventName] = make([]chan interface{}, 0)
		channels = p.events[eventName]
	}
	p.events[eventName] = append(channels, ch)
	p.stops[ch] = false

	go func() {
		for {
			v := <-ch
			if p.stops[ch] {
				break
			}
			h(v)
		}
	}()

	return ch
}

// Unsubscribe an event, by giving the chan object created when subscribing
func (p *PubSub) Unsubscribe(eventName string, ch chan interface{}) {
	channels, ok := p.events[eventName]
	if !ok {
		p.events[eventName] = make([]chan interface{}, 0)
		channels = p.events[eventName]
	}
	for i, c := range channels {
		if c == ch {
			// to stop subscirption, first set flag to true.
			// then notify the corresponding channel. It's going to continue
			// the routine, and found stop flag is true, then break from the loop
			p.stops[c] = true
			c <- true
			channels = append(channels[:i], channels[i+1:]...)
			p.events[eventName] = channels
			break
		}
	}
}

// publish an event by given name
func (p *PubSub) Publish(eventName string, msg interface{}) {
	channels, ok := p.events[eventName]
	if !ok {
		p.events[eventName] = make([]chan interface{}, 0)
		channels = p.events[eventName]
	}

	for _, c := range channels {
		c <- msg
	}
}
