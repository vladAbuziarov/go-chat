package message

import (
	"chatapp/internal/constants"
	messages_dto "chatapp/internal/dto/messages"
	chatEnts "chatapp/internal/entities/chat"
	"context"
	"database/sql"
	"fmt"

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

func (r *Repository) Create(ctx context.Context, message *chatEnts.Message) error {
	query := fmt.Sprintf(`
	INSERT INTO %s (conversation_id, sender_id, content, created_at)
	VALUES ($1, $2, $3, NOW()) RETURNING id`, constants.MessageTable)
	return r.db.GetContext(ctx, &message.ID, query, message.ConversationID, message.SenderID, message.Content)
}

func (r *Repository) Update(ctx context.Context, msg *chatEnts.Message, content string) error {
	query := fmt.Sprintf(`
	UPDATE %s SET content = $1, updated_at = NOW()
	WHERE id = $2
	RETURNING id, updated_at, content;`, constants.MessageTable)

	if err := r.db.QueryRowContext(ctx, query, content, msg.ID).Scan(&msg.ID, &msg.UpdatedAt, &msg.Content); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetMessages(ctx context.Context, qParams *messages_dto.GetMessageQueryParams) ([]*chatEnts.Message, error) {
	var messages []*chatEnts.Message
	query := r.getMessagesQueryString(qParams)
	args := []interface{}{qParams.ConvId, qParams.Limit}
	if qParams.LastReceivedId != nil {
		args = append(args, qParams.LastReceivedId)
	}
	if err := r.db.SelectContext(ctx, &messages, query, args...); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return messages, nil
}

func (r *Repository) GetMessageById(ctx context.Context, convId, msgId int64) (*chatEnts.Message, error) {
	query := fmt.Sprintf(`
	SELECT * FROM %s
	WHERE id = $1 AND conversation_id = $2
	LIMIT 1`, constants.MessageTable)
	message := &chatEnts.Message{}

	if err := r.db.GetContext(ctx, message, query, msgId, convId); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return message, nil
}

func (r *Repository) getMessagesQueryString(params *messages_dto.GetMessageQueryParams) string {
	whereQuery := "conversation_id = $1"
	if params.LastReceivedId != nil {
		whereQuery += " and id < $3"
	}
	return fmt.Sprintf(`
	SELECT * FROM %s WHERE %s
	ORDER BY (created_at, id) DESC
	LIMIT $2;`, constants.MessageTable, whereQuery)
}
