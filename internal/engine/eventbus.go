package engine

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

type EventBus struct {
	listeners    map[string]map[EventCallback]struct{}
	onceListeners map[EventCallback]struct{}
	queues       map[EventPriority][]queuedEvent
	processing   bool
}

func NewEventBus() *EventBus {
	return &EventBus{
		listeners:    make(map[string]map[EventCallback]struct{}),
		onceListeners: make(map[EventCallback]struct{}),
		queues:       make(map[EventPriority][]queuedEvent),
	}
}

func (eb *EventBus) On(event string, callback EventCallback) func() {
	if _, ok := eb.listeners[event]; !ok {
		eb.listeners[event] = make(map[EventCallback]struct{})
	}
	eb.listeners[event][callback] = struct{}{}
	return func() { eb.Off(event, callback) }
}

func (eb *EventBus) Once(event string, callback EventCallback) {
	eb.onceListeners[callback] = struct{}{}
	eb.On(event, callback)
}

func (eb *EventBus) Off(event string, callback EventCallback) {
	if listeners, ok := eb.listeners[event]; ok {
		delete(listeners, callback)
		delete(eb.onceListeners, callback)
	}
}

func (eb *EventBus) Emit(event string, data any) {
	listeners, ok := eb.listeners[event]
	if !ok {
		return
	}
	for cb := range listeners {
		cb(data)
		if _, isOnce := eb.onceListeners[cb]; isOnce {
			delete(eb.onceListeners, cb)
			delete(listeners, cb)
		}
	}
}

func (eb *EventBus) QueueEvent(event string, data any, priority EventPriority) {
	eb.queues[priority] = append(eb.queues[priority], queuedEvent{event, data, priority})
}

func (eb *EventBus) ProcessQueue() {
	if eb.processing {
		return
	}
	eb.processing = true

	order := []EventPriority{PriorityCritical, PriorityHigh, PriorityNormal, PriorityLow}
	for _, prio := range order {
		queue := eb.queues[prio]
		if len(queue) == 0 {
			continue
		}
		for len(queue) > 0 {
			qe := queue[0]
			queue = queue[1:]
			eb.Emit(qe.event, qe.data)
		}
		eb.queues[prio] = queue
	}

	eb.processing = false
}

func (eb *EventBus) Clear() {
	for k := range eb.listeners {
		delete(eb.listeners, k)
	}
	for k := range eb.onceListeners {
		delete(eb.onceListeners, k)
	}
	for k := range eb.queues {
		delete(eb.queues, k)
	}
}
