package conversation

import (
	"chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	eventlisteners "chatapp/internal/eventListeners"
	eventMocks "chatapp/internal/eventListeners/mocks"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	repoMocks "chatapp/internal/repositories/mocks"
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

type DummyLogger struct{}

func (d *DummyLogger) Info(msg string, args ...interface{})  {}
func (d *DummyLogger) Error(msg string, args ...interface{}) {}

func TestConversationService_CreateConversation_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}
	evlMock := eventMocks.NewMockIChatEventListener(ctrl)
	evls := eventlisteners.EventListeners{
		ChatEventListener: evlMock,
	}

	service := NewService(&repos, logger.NewLogger(), &evls)

	conv := &chat.Conversation{
		Name:    "Test Conversation",
		IsGroup: true,
	}
	participants := []users.UserId{users.UserId(1), users.UserId(2), users.UserId(3)}

	repoMock.
		EXPECT().
		Create(ctx, conv, participants).
		DoAndReturn(func(ctx context.Context, c *chat.Conversation, p []users.UserId) error {
			c.ID = 100
			return nil
		})

	createdConv, err := service.CreateConversation(ctx, conv.Name, conv.IsGroup, participants)
	t.Logf("%v", createdConv)
	assert.NoError(t, err)
	assert.NotNil(t, createdConv)
	assert.Equal(t, int64(100), createdConv.ID)
}

func TestConversationService_CreateConversation_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}
	evlMock := eventMocks.NewMockIChatEventListener(ctrl)
	evls := eventlisteners.EventListeners{
		ChatEventListener: evlMock,
	}

	service := NewService(&repos, logger.NewLogger(), &evls)

	conv := &chat.Conversation{
		Name:    "Test Conversation",
		IsGroup: false,
	}
	participants := []users.UserId{users.UserId(1), users.UserId(2)}

	repoMock.
		EXPECT().
		Create(ctx, conv, participants).
		Return(errors.New("db error"))

	createdConv, err := service.CreateConversation(ctx, conv.Name, conv.IsGroup, participants)
	assert.Error(t, err)
	assert.Nil(t, createdConv)
}

func TestConversationService_PostUserTyping_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}
	evlMock := eventMocks.NewMockIChatEventListener(ctrl)
	evls := eventlisteners.EventListeners{
		ChatEventListener: evlMock,
	}

	service := NewService(&repos, logger.NewLogger(), &evls)

	convID := int64(200)
	userID := users.UserId(1)

	repoMock.
		EXPECT().
		IsConversationExists(ctx, convID).
		Return(true, nil)
	repoMock.
		EXPECT().
		IsParticipant(ctx, convID, userID).
		Return(true, nil)

	evlMock.
		EXPECT().
		PostUserTyping(convID, userID).
		Return()

	err := service.PostUserTyping(ctx, userID, convID)
	assert.NoError(t, err)
}

func TestConversationService_PostUserTyping_AccessDenied(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}
	evlMock := eventMocks.NewMockIChatEventListener(ctrl)
	evls := eventlisteners.EventListeners{
		ChatEventListener: evlMock,
	}

	service := NewService(&repos, logger.NewLogger(), &evls)

	convID := int64(300)
	userID := users.UserId(1)

	repoMock.
		EXPECT().
		IsConversationExists(ctx, convID).
		Return(true, nil)
	repoMock.
		EXPECT().
		IsParticipant(ctx, convID, userID).
		Return(false, nil)

	err := service.PostUserTyping(ctx, userID, convID)
	assert.Error(t, err)
}
