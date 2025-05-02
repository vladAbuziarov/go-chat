package utils

import (
	"chatapp/internal/entities/users"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var (
	ErrConversationNotFound         = errors.New("conversation not found")
	ErrIsNotConversationParticipant = errors.New("user is not conversation participant")
	ErrFailedToCheckConversation    = errors.New("failed to check conversation")
	ErrFailedToCheckParticipant     = errors.New("faild to check participant")
)

type AccessChecker struct {
	repos  *repositories.Repositories
	logger logger.Logger
}

func NewAccesChecker(logger logger.Logger, repos *repositories.Repositories) *AccessChecker {
	return &AccessChecker{
		repos:  repos,
		logger: logger,
	}
}

func (a *AccessChecker) CanAccessConversation(ctx context.Context, cnvId int64, usrId users.UserId) error {
	cnvExists, err := a.repos.ConversationRepository.IsConversationExists(ctx, cnvId)
	if err != nil {
		a.logger.Error(ctx, fmt.Errorf("failed to check conversation by id: %w", err), slog.Int64("id", cnvId))
		return errors.Join(ErrFailedToCheckConversation, err)
	}
	if !cnvExists {
		return ErrConversationNotFound
	}

	isParticipant, err := a.repos.ConversationRepository.IsParticipant(ctx, cnvId, usrId)
	if err != nil {
		a.logger.Error(ctx, fmt.Errorf("failed check participant: %w", err), slog.Int64("id", cnvId), slog.Int64("userId", int64(usrId)))
		return errors.Join(ErrFailedToCheckParticipant, err)
	}
	if !isParticipant {
		return ErrIsNotConversationParticipant
	}
	return nil
}
