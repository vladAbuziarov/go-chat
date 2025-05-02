package fakes

import (
	"chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	"context"
	"errors"
	"sync"
	"time"
)

type FakeConversationRepository struct {
	mu            sync.Mutex
	conversations map[int64]*chat.Conversation
	participants  map[int64][]users.UserId
	nextID        int64
}

// NewFakeConversationRepository creates a new fake conversation repository.
func NewFakeConversationRepository() *FakeConversationRepository {
	return &FakeConversationRepository{
		conversations: make(map[int64]*chat.Conversation),
		participants:  make(map[int64][]users.UserId),
		nextID:        1,
	}
}

func (r *FakeConversationRepository) Create(ctx context.Context, conv *chat.Conversation, participants []users.UserId) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv.ID = r.nextID
	r.nextID++
	conv.CreatedAt = time.Now()
	conv.UpdatedAt = time.Now()
	r.conversations[conv.ID] = conv

	// Save participants, ensuring the conversation creator is included if not already.
	r.participants[conv.ID] = participants
	return nil
}

func (r *FakeConversationRepository) IsParticipant(ctx context.Context, convID int64, userID users.UserId) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ps, ok := r.participants[convID]
	if !ok {
		return false, errors.New("conversation not found")
	}
	for _, uid := range ps {
		if uid == users.UserId(userID) {
			return true, nil
		}
	}
	return false, nil
}

func (r *FakeConversationRepository) IsConversationExists(ctx context.Context, convID int64) (bool, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	_, ok := r.conversations[convID]
	return ok, nil
}

func (r *FakeConversationRepository) GetConversationById(ctx context.Context, convID int64) (*chat.Conversation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	conv, ok := r.conversations[convID]
	if !ok {
		return nil, nil
	}
	return conv, nil
}

func (r *FakeConversationRepository) GetParticipants(ctx context.Context, convID int64) ([]*chat.ConversationParticipant, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	ps, ok := r.participants[convID]
	if !ok {
		return nil, errors.New("conversation not found")
	}
	res := []*chat.ConversationParticipant{}
	for _, p := range ps {
		res = append(res, &chat.ConversationParticipant{
			ConversationID: convID,
			UserID:         int64(p),
			JoinedAt:       time.Now(),
		})
	}
	return res, nil
}
