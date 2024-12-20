package chat

import (
	chatEnts "chatapp/internal/entities/chat"
	"chatapp/internal/services/events"
	"sync"
)

type EventListener struct {
	ecs map[int64]*events.EventChannel
	mu  sync.RWMutex
}

func NewEventListener() *EventListener {
	return &EventListener{
		ecs: make(map[int64]*events.EventChannel),
	}
}

func (e *EventListener) SubscribeChannel(convId int64, l chan<- events.Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch, ok := e.ecs[convId]
	if !ok {
		ch = events.StartEventChannel()
		e.ecs[convId] = ch
	}
	ch.Subscribe(l)
}

func (e *EventListener) UnsubscribeChannel(convId int64, l chan<- events.Event) {
	e.mu.Lock()
	defer e.mu.Unlock()

	ch, ok := e.ecs[convId]
	if !ok {
		return
	}
	ch.Unsubscribe(l)

	if !ch.IsEmpty() {
		return
	}
	ch.Stop()
	delete(e.ecs, convId)
}

func (e *EventListener) PostUserTyping(cnvId, usrId int64) {
	ch, ok := e.ecs[cnvId]
	if !ok {
		return
	}
	ch.PostUserTyping(usrId)
}

func (e *EventListener) PostMessageCreated(msg *chatEnts.Message) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ch, ok := e.ecs[msg.ConversationID]
	if !ok {
		return
	}
	ch.PostMessageCreated(msg)
}

func (e *EventListener) PostMessageUpdated(msg *chatEnts.Message) {
	e.mu.RLock()
	defer e.mu.RUnlock()

	ch, ok := e.ecs[msg.ConversationID]
	if !ok {
		return
	}
	ch.PostMessageUpdated(msg)
}
