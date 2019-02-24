package watch

import (
	"sync"

	"github.com/gosoon/k8s-learning-notes/k8s-package/events/types"
)

// event chan queue
const incomingQueuLength = 100

type Broadcaster struct {
	lock             sync.Mutex
	incoming         chan types.Events
	watchers         map[int64]*broadcasterWatcher
	watchersQueue    int64
	watchQueueLength int64
	distributing     sync.WaitGroup
}

func NewBroadcaster(queueLength int64) *Broadcaster {
	m := &Broadcaster{
		incoming:         make(chan types.Events, incomingQueuLength),
		watchers:         map[int64]*broadcasterWatcher{},
		watchQueueLength: queueLength,
	}
	m.distributing.Add(1)
	go m.loop()
	return m
}

// receive event from producer
func (m *Broadcaster) Action(event *types.Events) {
	m.incoming <- *event
}

// send events to each watcher
func (m *Broadcaster) loop() {
	for event := range m.incoming {
		for _, w := range m.watchers {
			select {
			case w.result <- event:
			case <-w.stopped:
			default:
			}
		}
	}
	m.closeAll()
	m.distributing.Done()
}

// event broadcaster shutdown
func (m *Broadcaster) Shutdown() {
	close(m.incoming)
	m.distributing.Wait()
}

// when event broadcaster shutdown and disconnect all wacthers
func (m *Broadcaster) closeAll() {
	// TODO
	m.lock.Lock()
	defer m.lock.Unlock()
	for _, w := range m.watchers {
		close(w.result)
	}
	m.watchers = map[int64]*broadcasterWatcher{}
}

// one watcher stop receive event from broadcast
func (m *Broadcaster) stopWatching(id int64) {
	m.lock.Lock()
	defer m.lock.Unlock()
	w, ok := m.watchers[id]
	if !ok {
		return
	}
	delete(m.watchers, id)
	close(w.result)
}

// create a watcher
func (m *Broadcaster) Watch() Interface {
	watcher := &broadcasterWatcher{
		result:  make(chan types.Events, incomingQueuLength),
		stopped: make(chan struct{}),
		id:      m.watchQueueLength,
		m:       m,
	}
	m.watchers[m.watchersQueue] = watcher
	m.watchQueueLength++
	return watcher
}

// ------------
type Interface interface {
	Stop()
	ResultChan() <-chan types.Events
}

type broadcasterWatcher struct {
	result  chan types.Events
	stopped chan struct{}
	stop    sync.Once
	id      int64
	m       *Broadcaster
}

func (b *broadcasterWatcher) ResultChan() <-chan types.Events {
	return b.result
}

// when a watcher want disconnect and use Stop()
func (b *broadcasterWatcher) Stop() {
	b.stop.Do(func() {
		close(b.stopped)
		b.m.stopWatching(b.id)
	})
}
