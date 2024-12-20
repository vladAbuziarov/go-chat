package events

import (
	chatEnts "chatapp/internal/entities/chat"
	"sync"
	"sync/atomic"
)

type EventType string

const (
	EventTypeMessageCreated = "message_created"
	EventTypeMessageUpdated = "message_updated"

	EventTypeUserTyping = "user_typing"
)

type Event struct {
	Type EventType
	Data any
}

type EventChannel struct {
	listeners map[chan<- Event]struct{}
	msgsCh    chan Event

	mu            sync.RWMutex
	stopTriggered atomic.Bool
}

func StartEventChannel() *EventChannel {
	ch := &EventChannel{
		listeners: make(map[chan<- Event]struct{}),
		msgsCh:    make(chan Event, 100),
	}
	go ch.run()

	return ch
}

func (e *EventChannel) IsEmpty() bool {
	e.mu.RLock()
	defer e.mu.RUnlock()

	return len(e.listeners) == 0
}

func (e *EventChannel) Stop() {
	e.stopTriggered.Store(true)
}

func (e *EventChannel) PostMessageCreated(msg *chatEnts.Message) {
	e.msgsCh <- Event{
		Type: EventTypeMessageCreated,
		Data: msg,
	}
}
func (e *EventChannel) PostMessageUpdated(msg *chatEnts.Message) {
	e.msgsCh <- Event{
		Type: EventTypeMessageUpdated,
		Data: msg,
	}
}

func (e *EventChannel) PostUserTyping(usrId int64) {
	e.msgsCh <- Event{
		Type: EventTypeUserTyping,
		Data: map[string]int64{"user_id": usrId},
	}
}

func (e *EventChannel) Subscribe(l chan<- Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.listeners[l] = struct{}{}
}

func (e *EventChannel) Unsubscribe(l chan<- Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	delete(e.listeners, l)
}

func (e *EventChannel) run() {
	for msg := range e.msgsCh {
		if e.stopTriggered.Load() {
			close(e.msgsCh)
			return
		}
		e.mu.RLock()
		for l := range e.listeners {
			l <- msg
		}
		e.mu.RUnlock()
	}
}
