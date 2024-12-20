package services

import (
	"chatapp/internal/clients"
	"chatapp/internal/config"
	messages_dto "chatapp/internal/dto/messages"
	user_dto "chatapp/internal/dto/user"
	"chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	eventlisteners "chatapp/internal/eventListeners"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	user "chatapp/internal/services/auth"
	"chatapp/internal/services/auth/jwt"
	"chatapp/internal/services/chat/conversation"
	"chatapp/internal/services/chat/message"
	"context"
)

type JWTServiceInterface interface {
	CreateToken(ctx context.Context, userID int64) (string, error)
	VerifyAuthToken(token string) (*int64, error)
}
type UserServiceInterface interface {
	Register(ctx context.Context, user *user_dto.CreateUserDTO) (*users.User, error)
	Login(ctx context.Context, email, password string) (*users.User, error)
}
type ConversationServiceInterface interface {
	CreateConversation(ctx context.Context, name string, isGroup bool, participantIDs []int64) (*chat.Conversation, error)
	PostUserTyping(ctx context.Context, usrId, cnvId int64) error
}
type MessageServiceInterface interface {
	SendMessage(ctx context.Context, conversationID, senderID int64, content string) (*chat.Message, error)
	GetMessages(ctx context.Context, params *messages_dto.GetMessageQueryParams) ([]*chat.Message, error)
	UpdateMessage(ctx context.Context, conversationID, messageID, userID int64, content string) (*chat.Message, error)
}
type Services struct {
	UserService         UserServiceInterface
	JwtService          JWTServiceInterface
	ConversationService ConversationServiceInterface
	MessageService      MessageServiceInterface
}

func NewServices(cfg *config.Config, cls *clients.Clients, logger logger.Logger, repos *repositories.Repositories, evls *eventlisteners.EventListeners) *Services {
	return &Services{
		UserService:         user.NewService(logger, repos),
		JwtService:          jwt.NewService(cfg, logger),
		ConversationService: conversation.NewService(cls, repos, logger, evls.ChatEventListener),
		MessageService:      message.NewService(logger, repos, evls.ChatEventListener),
	}
}
