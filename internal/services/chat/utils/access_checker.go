package utils

import (
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

func (a *AccessChecker) CanAccessConversation(ctx context.Context, cnvId, usrId int64) error {
	cnvExists, err := a.repos.ConversationRepository.IsConversationExists(ctx, cnvId)
	if err != nil {
		a.logger.Error(ctx, fmt.Errorf("failed to check conversation by id: %w", err), slog.Int64("id", cnvId))
		return fmt.Errorf("fail to check coversation: %w", err)
	}
	if !cnvExists {
		return ErrConversationNotFound
	}

	isParticipant, err := a.repos.ConversationRepository.IsParticipant(ctx, cnvId, usrId)
	if err != nil {
		a.logger.Error(ctx, fmt.Errorf("failed check participant: %w", err), slog.Int64("id", cnvId), slog.Int64("userId", usrId))
		return fmt.Errorf("faild to check participant: %w", err)
	}
	if !isParticipant {
		return ErrIsNotConversationParticipant
	}
	return nil
}
