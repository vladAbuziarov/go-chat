package chat

import (
	"chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	"chatapp/internal/services/events"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConversationEventListener_PostUserTyping(t *testing.T) {
	listener := NewEventListener()
	convID := int64(123)

	ch := make(chan events.Event, 1)

	listener.SubscribeChannel(convID, ch)

	listener.PostUserTyping(convID, 42)

	select {
	case ev := <-ch:
		assert.Equal(t, events.EventType(events.EventTypeUserTyping), ev.Type)
		assert.Equal(t, map[string]users.UserId{"user_id": users.UserId(42)}, ev.Data)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for user typing event")
	}
}

func TestConversationEventListener_PostMessageCreated(t *testing.T) {
	listener := NewEventListener()
	convID := int64(123)

	ch := make(chan events.Event, 1)

	listener.SubscribeChannel(convID, ch)

	msg := &chat.Message{
		ID:             int64(1),
		ConversationID: convID,
		SenderID:       users.UserId(1),
		Content:        "test",
	}
	listener.PostMessageCreated(msg)

	select {
	case ev := <-ch:
		assert.Equal(t, events.EventType(events.EventTypeMessageCreated), ev.Type)
		assert.Equal(t, msg, ev.Data)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for message created event")
	}
}

func TestConversationEventListener_PostMessageUpdated(t *testing.T) {
	listener := NewEventListener()
	convID := int64(123)

	ch := make(chan events.Event, 1)

	listener.SubscribeChannel(convID, ch)

	msg := &chat.Message{
		ID:             int64(1),
		ConversationID: convID,
		SenderID:       users.UserId(1),
		Content:        "test",
	}
	listener.PostMessageUpdated(msg)

	select {
	case ev := <-ch:
		assert.Equal(t, events.EventType(events.EventTypeMessageUpdated), ev.Type)
		assert.Equal(t, msg, ev.Data)
	case <-time.After(1 * time.Second):
		t.Fatal("timed out waiting for message updated event")
	}
}
func TestConversationEventListener_Unsubscribe(t *testing.T) {
	listener := NewEventListener()
	convID := int64(123)
	ch := make(chan events.Event, 1)

	listener.SubscribeChannel(convID, ch)
	listener.UnsubscribeChannel(convID, ch)

	listener.PostUserTyping(convID, 42)

	select {
	case ev := <-ch:
		t.Fatalf("unexpectedly received event: %+v", ev)
	case <-time.After(500 * time.Millisecond):
	}
}

func TestConversationEventListener_MultipleSubscribers(t *testing.T) {
	listener := NewEventListener()
	convID := int64(123)
	ch1 := make(chan events.Event, 1)
	ch2 := make(chan events.Event, 1)

	listener.SubscribeChannel(convID, ch1)
	listener.SubscribeChannel(convID, ch2)

	listener.PostUserTyping(convID, 42)

	select {
	case ev := <-ch1:
		assert.Equal(t, events.EventType(events.EventTypeUserTyping), ev.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ch1 did not receive event")
	}
	select {
	case ev := <-ch2:
		assert.Equal(t, events.EventType(events.EventTypeUserTyping), ev.Type)
	case <-time.After(500 * time.Millisecond):
		t.Fatal("ch2 did not receive event")
	}
}
