package utils

import (
	"chatapp/internal/entities/users"
	repoMocks "chatapp/internal/repositories/mocks"
	"errors"

	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestAccessChecker_ConversationNotFound_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}

	ac := NewAccesChecker(logger.NewLogger(), &repos)

	repoMock.
		EXPECT().
		IsConversationExists(ctx, int64(1)).
		Return(false, nil)

	err := ac.CanAccessConversation(ctx, int64(1), users.UserId(1))
	assert.ErrorIs(t, err, ErrConversationNotFound)
}

func TestAccessChecker_CanAccessConversation_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}

	ac := NewAccesChecker(logger.NewLogger(), &repos)

	repoMock.
		EXPECT().
		IsConversationExists(ctx, int64(1)).
		Return(false, errors.New("db error"))

	err := ac.CanAccessConversation(ctx, int64(1), users.UserId(1))
	assert.Error(t, err)
	assert.ErrorIs(t, err, ErrFailedToCheckConversation)
}

func TestAccessChecker_IsNotParticipantError_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}

	ac := NewAccesChecker(logger.NewLogger(), &repos)

	repoMock.
		EXPECT().
		IsConversationExists(ctx, int64(1)).
		Return(true, nil)

	repoMock.
		EXPECT().
		IsParticipant(ctx, int64(1), users.UserId(1)).
		Return(false, nil)

	err := ac.CanAccessConversation(ctx, int64(1), users.UserId(1))
	assert.ErrorIs(t, err, ErrIsNotConversationParticipant)
}

func TestAccessChecker_ParticipantCheck_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}

	ac := NewAccesChecker(logger.NewLogger(), &repos)

	repoMock.
		EXPECT().
		IsConversationExists(ctx, int64(1)).
		Return(true, nil)

	repoMock.
		EXPECT().
		IsParticipant(ctx, int64(1), users.UserId(1)).
		Return(true, errors.New("db error"))

	err := ac.CanAccessConversation(ctx, int64(1), users.UserId(1))
	assert.ErrorIs(t, err, ErrFailedToCheckParticipant)
}

func TestAccessChecker_CanAccessConversationSuccess(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	repoMock := repoMocks.NewMockConversationRepositoryInterface(ctrl)
	repos := repositories.Repositories{
		ConversationRepository: repoMock,
	}

	ac := NewAccesChecker(logger.NewLogger(), &repos)

	repoMock.
		EXPECT().
		IsConversationExists(ctx, int64(1)).
		Return(true, nil)

	repoMock.
		EXPECT().
		IsParticipant(ctx, int64(1), users.UserId(1)).
		Return(true, nil)

	err := ac.CanAccessConversation(ctx, int64(1), users.UserId(1))
	assert.NoError(t, err)
	assert.Nil(t, err)
}
