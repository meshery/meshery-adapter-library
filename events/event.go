package events

import (
	"fmt"
	"sync"
)

type EventBuffer struct {
	circularqueue  []interface{}
	clientChannels []chan interface{}
	last           int //last=-1 means queue is empty
	mx             sync.Mutex
	clmx           sync.Mutex
}

//TODO: Persist older than 10 events in non-volatile db for obtaining history of operations
// This is the maximum number of events that will be stored in-memory in each adapter.

var maxElements = 10

func NewEventBuffer() *EventBuffer {
	return &EventBuffer{
		circularqueue: make([]interface{}, maxElements),
		last:          0,
	}
}

// Each event will be first stored in a circular queue to return the last 10 operations whenever a new client connects.
// After that, further events will be pushed to the client.
func (e *EventBuffer) Enqueue(i interface{}) {
	e.mx.Lock()
	defer e.mx.Unlock()
	pos := e.last % len(e.circularqueue)
	fmt.Println("pushing data to queue at pos", pos)
	e.circularqueue[pos] = i
	e.last++
	go func() {
		for _, ch := range e.clientChannels {
			ch <- i
		}
	}()
}
func (e *EventBuffer) Copy(client chan interface{}) {
	e.mx.Lock()
	defer e.mx.Unlock()
	var events []interface{}
	client = make(chan interface{}, 10)
	for i := 0; i < e.last%len(e.circularqueue); i++ {
		ev := e.circularqueue[i]
		events = append(events, ev)
		go func(ev interface{}) {
			fmt.Println("copying data to new client ", ev)
			client <- ev
		}(ev)
	}
	return
}

func (e *EventBuffer) Subscribe(ch chan interface{}) {
	e.clmx.Lock()
	defer e.clmx.Unlock()
	e.clientChannels = append(e.clientChannels, ch)
}
