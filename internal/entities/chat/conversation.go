package chat

import "time"

// Conversation represents a chat conversation between users.
// @Description Represents a chat conversation between users.
// @Swagger:ignore This tag is optional if you want to include/exclude the struct.
// @Tags conversations
type Conversation struct {
	ID        int64     `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	IsGroup   bool      `db:"is_group" json:"is_group"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
	UpdatedAt time.Time `db:"updated_at" json:"updated_at"`
}

type ConversationParticipant struct {
	ConversationID int64     `db:"conversation_id" json:"conversation_id"`
	UserID         int64     `db:"user_id" json:"user_id"`
	JoinedAt       time.Time `db:"joined_at" json:"joined_at"`
}
