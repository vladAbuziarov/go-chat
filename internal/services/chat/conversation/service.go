package conversation

import (
	chatEnts "chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	eventlisteners "chatapp/internal/eventListeners"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/services/chat/utils"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var (
	ErrIsNotConversationParticipant = errors.New("user is not conversation participant")
)

type Service struct {
	repos  *repositories.Repositories
	aCh    *utils.AccessChecker
	logger logger.Logger

	evls *eventlisteners.EventListeners
}

func NewService(
	repos *repositories.Repositories,
	logger logger.Logger,
	evls *eventlisteners.EventListeners,
) *Service {
	return &Service{
		repos:  repos,
		aCh:    utils.NewAccesChecker(logger, repos),
		logger: logger,
		evls:   evls,
	}
}

func (s *Service) CreateConversation(ctx context.Context, name string, isGroup bool, pts []users.UserId) (cnv *chatEnts.Conversation, err error) {
	cnv = &chatEnts.Conversation{
		Name:    name,
		IsGroup: isGroup,
	}

	if err = s.repos.ConversationRepository.Create(ctx, cnv, pts); err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed create conversation: %w", err),
			slog.Any("conversation", cnv))
		return nil, fmt.Errorf("failed create conversation: %w", err)
	}

	s.logger.Info(ctx, "conversation created", slog.Any("conversation", cnv))
	return cnv, nil
}

func (s *Service) PostUserTyping(ctx context.Context, usrId users.UserId, cnvId int64) error {
	if err := s.aCh.CanAccessConversation(ctx, cnvId, usrId); err != nil {
		s.logger.Error(ctx, fmt.Errorf("PostUserTyping: failed to access conversation: %w", err), slog.Int64("conversation", cnvId), slog.Int64("user", int64(usrId)))
		return err
	}
	s.evls.ChatEventListener.PostUserTyping(cnvId, usrId)
	return nil
}
