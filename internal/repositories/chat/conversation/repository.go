package conversation

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"chatapp/internal/constants"
	chatEnts "chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	"chatapp/internal/logger"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db     *sqlx.DB
	logger logger.Logger
}

func NewRepository(db *sqlx.DB, logger logger.Logger) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}

func (r *Repository) Create(ctx context.Context, cnv *chatEnts.Conversation, pts []users.UserId) error {
	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{
		Isolation: sql.LevelReadCommitted,
		ReadOnly:  false,
	})
	if err != nil {
		return fmt.Errorf("failed begin transaction: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(); rollbackErr != nil && errors.Is(rollbackErr, sql.ErrTxDone) {
			r.logger.Error(ctx, fmt.Errorf("failed to rollback transaction: %w", rollbackErr))
		}
	}()

	query := fmt.Sprintf(`
	INSERT INTO %s (name, is_group, created_at, updated_at)
	VALUES ($1, $2, NOW(), NOW())
	RETURNING id`, constants.ConversationTable)
	err = tx.GetContext(ctx, &cnv.ID, query, cnv.Name, cnv.IsGroup)

	for _, p := range pts {
		pQuery := fmt.Sprintf(`
		INSERT INTO %s (conversation_id, user_id, joined_at)
		VALUES ($1, $2, NOW())`, constants.ConversationParticipantTable)

		_, err = tx.ExecContext(ctx, pQuery, cnv.ID, p)
	}

	return err
}

func (r *Repository) IsParticipant(ctx context.Context, cnvId int64, usrId users.UserId) (bool, error) {
	var exists bool
	query := fmt.Sprintf(`
	SELECT EXISTS(
	SELECT 1 FROM %s WHERE conversation_id = $1 AND user_id = $2)`, constants.ConversationParticipantTable)

	if err := r.db.GetContext(ctx, &exists, query, cnvId, usrId); err != nil {
		return false, err
	}
	return exists, nil
}
func (r *Repository) IsConversationExists(ctx context.Context, cnvId int64) (bool, error) {
	var exists bool
	query := fmt.Sprintf(`
	SELECT EXISTS(
	SELECT 1 FROM %s WHERE id = $1)`, constants.ConversationTable)

	if err := r.db.GetContext(ctx, &exists, query, cnvId); err != nil {
		return false, err
	}
	return exists, nil
}

func (r *Repository) GetConversationById(ctx context.Context, cnvId int64) (*chatEnts.Conversation, error) {
	conv := &chatEnts.Conversation{}
	query := fmt.Sprintf(`
	SELECT * FROM %s WHERE id = $1 LIMIT 1`, constants.ConversationTable)

	if err := r.db.GetContext(ctx, conv, query, cnvId); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return conv, nil
}
func (r *Repository) GetParticipants(ctx context.Context, cnvId int64) ([]*chatEnts.ConversationParticipant, error) {
	var pts []*chatEnts.ConversationParticipant
	query := fmt.Sprintf(`
	SELECT FROM %s WHERE conversation_id = $1`, constants.ConversationParticipantTable)

	err := r.db.SelectContext(ctx, &pts, query, cnvId)
	return pts, err
}
