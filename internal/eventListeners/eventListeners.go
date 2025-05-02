package eventlisteners

import (
	chatEnts "chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	"chatapp/internal/services/chat"
	"chatapp/internal/services/events"
)

type IChatEventListener interface {
	SubscribeChannel(convId int64, l chan<- events.Event)
	UnsubscribeChannel(convId int64, l chan<- events.Event)
	PostUserTyping(cnvId int64, usrId users.UserId)
	PostMessageCreated(msg *chatEnts.Message)
	PostMessageUpdated(msg *chatEnts.Message)
}

type EventListeners struct {
	ChatEventListener IChatEventListener
}

func NewEventListeners() *EventListeners {
	return &EventListeners{
		ChatEventListener: chat.NewEventListener(),
	}
}
