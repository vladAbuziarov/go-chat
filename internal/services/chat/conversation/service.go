package conversation

import (
	"chatapp/internal/clients"
	chatEnts "chatapp/internal/entities/chat"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	chat_events "chatapp/internal/services/chat"
	"chatapp/internal/services/chat/utils"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jmoiron/sqlx"
)

var (
	ErrIsNotConversationParticipant = errors.New("user is not conversation participant")
)

type Service struct {
	repos  *repositories.Repositories
	aCh    *utils.AccessChecker
	logger logger.Logger
	db     *sqlx.DB

	evl *chat_events.EventListener
}

func NewService(
	clients *clients.Clients,
	repos *repositories.Repositories,
	logger logger.Logger,
	evl *chat_events.EventListener,
) *Service {
	return &Service{
		repos:  repos,
		aCh:    utils.NewAccesChecker(logger, repos),
		logger: logger,
		db:     clients.Postgres,
		evl:    evl,
	}
}

func (s *Service) CreateConversation(ctx context.Context, name string, isGroup bool, pts []int64) (cnv *chatEnts.Conversation, err error) {
	cnv = &chatEnts.Conversation{
		Name:    name,
		IsGroup: isGroup,
	}
	tx, err := s.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed begin transaction: %w", err)
	}
	defer func() {
		if p := recover(); p != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.logger.Error(ctx, fmt.Errorf("failed to rollback transaction after panic: %w", rollbackErr))
			}
			panic(p)
		} else if err != nil {
			if rollbackErr := tx.Rollback(); rollbackErr != nil {
				s.logger.Error(ctx, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
			}
		} else {
			if commitErr := tx.Commit(); commitErr != nil {
				err = fmt.Errorf("failed to commit transaction: %w", commitErr)
			}
		}
	}()

	if err = s.repos.ConversationRepository.Create(ctx, tx, cnv); err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed create conversation: %w", err),
			slog.Any("conversation", cnv))
		return nil, fmt.Errorf("failed create conversation: %w", err)
	}
	for _, p := range pts {
		if err = s.repos.ConversationRepository.AddParticipant(ctx, tx, cnv.ID, p); err != nil {
			s.logger.Error(ctx, fmt.Errorf("faild to add participant in conversation: %w", err),
				slog.Any("conversation", cnv), slog.Int64("userId", p))
			return nil, fmt.Errorf("faild to add participant in conversation: %w", err)
		}
	}

	s.logger.Info(ctx, "conversation created", slog.Any("conversation", cnv))
	return cnv, nil
}

func (s *Service) PostUserTyping(ctx context.Context, usrId, cnvId int64) error {
	if err := s.aCh.CanAccessConversation(ctx, cnvId, usrId); err != nil {
		s.logger.Error(ctx, fmt.Errorf("PostUserTyping: failed to access conversation: %w", err), slog.Int64("conversation", cnvId), slog.Int64("user", usrId))
		return err
	}
	s.evl.PostUserTyping(cnvId, usrId)
	return nil
}
