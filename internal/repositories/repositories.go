package repositories

import (
	"chatapp/internal/clients"
	messages_dto "chatapp/internal/dto/messages"
	user_dto "chatapp/internal/dto/user"
	"chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	"chatapp/internal/logger"
	"chatapp/internal/repositories/chat/conversation"
	"chatapp/internal/repositories/chat/message"
	"chatapp/internal/repositories/user"
	"context"
)

type UserRepositoryInterface interface {
	Create(ctx context.Context, userDto *user_dto.CreateUserDTO) (*users.User, error)
	GetUserById(ctx context.Context, id users.UserId) (*users.User, error)
	GetUserByEmail(ctx context.Context, email string) (*users.User, error)
}
type ConversationRepositoryInterface interface {
	Create(ctx context.Context, cnv *chat.Conversation, pts []users.UserId) error
	IsParticipant(ctx context.Context, cnvId int64, usrId users.UserId) (bool, error)
	IsConversationExists(ctx context.Context, cnvId int64) (bool, error)
	GetConversationById(ctx context.Context, cnvId int64) (*chat.Conversation, error)
	GetParticipants(ctx context.Context, cnvId int64) ([]*chat.ConversationParticipant, error)
}
type MessageRepositoryInterface interface {
	Create(ctx context.Context, message *chat.Message) error
	GetMessages(ctx context.Context, qParams *messages_dto.GetMessageQueryParams) ([]*chat.Message, error)
	Update(ctx context.Context, msg *chat.Message, content string) error
	GetMessageById(ctx context.Context, convId, msgId int64) (*chat.Message, error)
}
type Repositories struct {
	UserRepository         UserRepositoryInterface
	ConversationRepository ConversationRepositoryInterface
	MessageRepository      MessageRepositoryInterface
}

func NewRepositories(clients *clients.Clients, logger logger.Logger) *Repositories {
	return &Repositories{
		UserRepository:         user.NewRepository(clients.Postgres),
		ConversationRepository: conversation.NewRepository(clients.Postgres, logger),
		MessageRepository:      message.NewRepository(clients.Postgres),
	}
}
