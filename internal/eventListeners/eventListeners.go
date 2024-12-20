package eventlisteners

import "chatapp/internal/services/chat"

type EventListeners struct {
	ChatEventListener *chat.EventListener
}

func NewEventListeners() *EventListeners {
	return &EventListeners{
		ChatEventListener: chat.NewEventListener(),
	}
}
