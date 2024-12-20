package chat

import "time"

type Message struct {
	ID             int64     `db:"id" json:"id"`
	ConversationID int64     `db:"conversation_id" json:"conversation_id"`
	SenderID       int64     `db:"sender_id" json:"sender_id"`
	Content        string    `db:"content" json:"content"`
	CreatedAt      time.Time `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}
