package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/gosoon/k8s-learning-notes/k8s-package/events/types"
	"github.com/gosoon/k8s-learning-notes/k8s-package/events/watch"
)

// watchers start id
const queueLength = int64(1)

// EventBroadcaster xxx
type EventBroadcaster interface {
	Eventf(etype, reason, message string)
	StartLogging() watch.Interface
	Stop()
}

// eventBroadcaster xxx
type eventBroadcasterImpl struct {
	*watch.Broadcaster
}

func NewEventBroadcaster() EventBroadcaster {
	return &eventBroadcasterImpl{watch.NewBroadcaster(queueLength)}
}

func (eventBroadcaster *eventBroadcasterImpl) Stop() {
	eventBroadcaster.Shutdown()
}

// generate event
func (eventBroadcaster *eventBroadcasterImpl) Eventf(etype, reason, message string) {
	events := &types.Events{Type: etype, Reason: reason, Message: message}
	// send event to broadcast
	eventBroadcaster.Action(events)
}

// register a watcher and receive event
func (eventBroadcaster *eventBroadcasterImpl) StartLogging() watch.Interface {
	watcher := eventBroadcaster.Watch()
	go func() {
		for watchEvent := range watcher.ResultChan() {
			fmt.Printf("%v\n", watchEvent)
		}
	}()

	// test watcher client stop
	go func() {
		time.Sleep(time.Second * 4)
		watcher.Stop()
	}()

	return watcher
}

func main() {
	eventBroadcast := NewEventBroadcaster()

	var wg sync.WaitGroup
	wg.Add(1)
	// producer event
	go func() {
		defer wg.Done()
		time.Sleep(time.Second)
		eventBroadcast.Eventf("add", "test", "1")
		time.Sleep(time.Second * 2)
		eventBroadcast.Eventf("add", "test", "2")
		time.Sleep(time.Second * 3)
		eventBroadcast.Eventf("add", "test", "3")
		// eventBroadcast.Stop() is test event producer stop
		//eventBroadcast.Stop()
	}()

	eventBroadcast.StartLogging()
	wg.Wait()
}
