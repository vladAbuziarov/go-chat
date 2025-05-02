package services

import (
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
	"chatapp/internal/services/hash"
	"context"
	"time"
)

type JWTServiceInterface interface {
	CreateToken(ctx context.Context, userID users.UserId) (string, error)
	VerifyAuthToken(token string) (*int64, error)
}
type UserServiceInterface interface {
	Register(ctx context.Context, user *user_dto.CreateUserDTO) (*users.User, error)
	Login(ctx context.Context, email, password string) (*users.User, error)
}
type ConversationServiceInterface interface {
	CreateConversation(ctx context.Context, name string, isGroup bool, participantIDs []users.UserId) (*chat.Conversation, error)
	PostUserTyping(ctx context.Context, usrId users.UserId, cnvId int64) error
}
type MessageServiceInterface interface {
	SendMessage(ctx context.Context, conversationID int64, senderID users.UserId, content string) (*chat.Message, error)
	GetMessages(ctx context.Context, params *messages_dto.GetMessageQueryParams) ([]*chat.Message, error)
	UpdateMessage(ctx context.Context, conversationID, messageID int64, userID users.UserId, content string) (*chat.Message, error)
}
type Services struct {
	UserService         UserServiceInterface
	JwtService          JWTServiceInterface
	ConversationService ConversationServiceInterface
	MessageService      MessageServiceInterface
}

func NewServices(cfg *config.Config, logger logger.Logger, repos *repositories.Repositories, evls *eventlisteners.EventListeners) *Services {
	return &Services{
		UserService:         user.NewService(logger, hash.NewService(), repos),
		JwtService:          jwt.NewService(5*time.Hour, cfg, logger),
		ConversationService: conversation.NewService(repos, logger, evls),
		MessageService:      message.NewService(logger, repos, evls),
	}
}
