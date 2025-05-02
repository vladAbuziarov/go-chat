package message

import (
	messages_dto "chatapp/internal/dto/messages"
	chatEnts "chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	eventlisteners "chatapp/internal/eventListeners"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/services/chat/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
)

var (
	ErrMessageNotFound = errors.New("message not found")
)

type Service struct {
	repos  *repositories.Repositories
	aCh    *utils.AccessChecker
	logger logger.Logger

	evls *eventlisteners.EventListeners
}

func NewService(logger logger.Logger, repos *repositories.Repositories, evls *eventlisteners.EventListeners) *Service {
	return &Service{
		repos:  repos,
		aCh:    utils.NewAccesChecker(logger, repos),
		logger: logger,

		evls: evls,
	}
}

func (s *Service) SendMessage(ctx context.Context, cnvId int64, sederId users.UserId, content string) (*chatEnts.Message, error) {
	if err := s.aCh.CanAccessConversation(ctx, cnvId, sederId); err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed create message: %w", err), slog.Int64("id", cnvId))
		return nil, err
	}
	message := &chatEnts.Message{
		ConversationID: cnvId,
		SenderID:       sederId,
		Content:        content,
	}

	if err := s.repos.MessageRepository.Create(ctx, message); err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed create message: %w", err), slog.Any("message", message))
		return nil, fmt.Errorf("failed create message: %w", err)
	}
	s.evls.ChatEventListener.PostMessageCreated(message)
	s.logger.Info(ctx, "message created", slog.Any("message", message))
	return message, nil
}

func (s *Service) UpdateMessage(ctx context.Context, cnvId, msgId int64, usrId users.UserId, content string) (*chatEnts.Message, error) {
	if err := s.aCh.CanAccessConversation(ctx, cnvId, usrId); err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed update message: %w", err), slog.Int64("id", cnvId))
		return nil, err
	}
	message, err := s.repos.MessageRepository.GetMessageById(ctx, cnvId, msgId)
	if err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed to get message for update: %w", err), slog.Int64("conversation", cnvId), slog.Int64("message", msgId), slog.String("content", content))
		return nil, fmt.Errorf("failed to update message: %w", err)
	}
	if message == nil {
		return nil, ErrMessageNotFound
	}
	if message.SenderID != usrId {
		return nil, fmt.Errorf("have not access to update message content")
	}
	if err := s.repos.MessageRepository.Update(ctx, message, content); err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed to update message: %w", err), slog.Any("message", message), slog.String("content", content))
		return nil, fmt.Errorf("failed to update message: %w", err)
	}

	s.evls.ChatEventListener.PostMessageUpdated(message)
	s.logger.Info(ctx, "message updated", slog.Any("message", message))
	return message, nil
}

func (s *Service) GetMessages(ctx context.Context, params *messages_dto.GetMessageQueryParams) ([]*chatEnts.Message, error) {
	if err := s.aCh.CanAccessConversation(ctx, params.ConvId, params.UserId); err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed get message list: %w", err), slog.Any("params", params))
		return nil, err
	}

	messages, err := s.repos.MessageRepository.GetMessages(ctx, params)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		s.logger.Error(ctx, fmt.Errorf("failed get messages: %w", err), slog.Any("params", params))
		return nil, fmt.Errorf("failed get messages: %w", err)
	}
	return messages, nil
}
