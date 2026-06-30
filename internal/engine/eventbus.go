package engine

import "sync"

type EventPriority int

const (
	PriorityCritical EventPriority = 0
	PriorityHigh     EventPriority = 1
	PriorityNormal   EventPriority = 2
	PriorityLow      EventPriority = 3
)

type queuedEvent struct {
	event    string
	data     any
	priority EventPriority
}

type EventCallback func(data any)

type listener struct {
	callback EventCallback
	once     bool
	id       uint64
}

type EventBus struct {
	mu         sync.Mutex
	listeners  map[string][]listener
	nextID     uint64
	queues     map[EventPriority][]queuedEvent
	processing bool
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners: make(map[string][]listener),
		queues:    make(map[EventPriority][]queuedEvent),
	}
}

func (eb *EventBus) On(event string, callback EventCallback) func() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	id := eb.nextID
	eb.nextID++
	eb.listeners[event] = append(eb.listeners[event], listener{callback: callback, once: false, id: id})
	return func() { eb.remove(event, id) }
}

func (eb *EventBus) Once(event string, callback EventCallback) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	id := eb.nextID
	eb.nextID++
	eb.listeners[event] = append(eb.listeners[event], listener{callback: callback, once: true, id: id})
}

func (eb *EventBus) remove(event string, id uint64) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	listeners := eb.listeners[event]
	for i, l := range listeners {
		if l.id == id {
			eb.listeners[event] = append(listeners[:i], listeners[i+1:]...)
			return
		}
	}
}

func (eb *EventBus) Emit(event string, data any) {
	eb.mu.Lock()
	listeners := make([]listener, len(eb.listeners[event]))
	copy(listeners, eb.listeners[event])
	eb.mu.Unlock()

	var toRemove []uint64
	for _, l := range listeners {
		l.callback(data)
		if l.once {
			toRemove = append(toRemove, l.id)
		}
	}

	if len(toRemove) > 0 {
		eb.mu.Lock()
		remaining := eb.listeners[event][:0]
		for _, l := range eb.listeners[event] {
			remove := false
			for _, id := range toRemove {
				if l.id == id {
					remove = true
					break
				}
			}
			if !remove {
				remaining = append(remaining, l)
			}
		}
		eb.listeners[event] = remaining
		eb.mu.Unlock()
	}
}

func (eb *EventBus) QueueEvent(event string, data any, priority EventPriority) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.queues[priority] = append(eb.queues[priority], queuedEvent{event, data, priority})
}

func (eb *EventBus) ProcessQueue() {
	eb.mu.Lock()
	if eb.processing {
		eb.mu.Unlock()
		return
	}
	eb.processing = true

	allQueues := make([][]queuedEvent, 4)
	for i := PriorityCritical; i <= PriorityLow; i++ {
		allQueues[i] = eb.queues[i]
		delete(eb.queues, i)
	}
	eb.processing = false
	eb.mu.Unlock()

	order := []EventPriority{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}
	for _, prio := range order {
		for _, qe := range allQueues[prio] {
			eb.Emit(qe.event, qe.data)
		}
	}
}

func (eb *EventBus) Clear() {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	for k := range eb.listeners {
		delete(eb.listeners, k)
	}
	for k := range eb.queues {
		delete(eb.queues, k)
	}
}
