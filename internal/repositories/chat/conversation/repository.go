package conversation

import (
	"context"
	"database/sql"
	"fmt"

	"chatapp/internal/constants"
	chatEnts "chatapp/internal/entities/chat"

	"github.com/jmoiron/sqlx"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		db: db,
	}
}

func (r *Repository) Create(ctx context.Context, tx *sqlx.Tx, cnv *chatEnts.Conversation) error {
	query := fmt.Sprintf(`
	INSERT INTO %s (name, is_group, created_at, updated_at)
	VALUES ($1, $2, NOW(), NOW())
	RETURNING id`, constants.ConversationTable)

	return tx.GetContext(ctx, &cnv.ID, query, cnv.Name, cnv.IsGroup)
}

func (r *Repository) AddParticipant(ctx context.Context, tx *sqlx.Tx, cnvId, usrId int64) error {
	query := fmt.Sprintf(`
	INSERT INTO %s (conversation_id, user_id, joined_at)
	VALUES ($1, $2, NOW())`, constants.ConversationParticipantTable)

	_, err := tx.ExecContext(ctx, query, cnvId, usrId)
	return err
}

func (r *Repository) IsParticipant(ctx context.Context, cnvId, usrId int64) (bool, error) {
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
