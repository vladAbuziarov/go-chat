package handlers

import (
	"chatapp/cmd/server/handlers/auth"
	"chatapp/cmd/server/handlers/chat/conversation"
	"chatapp/cmd/server/handlers/chat/message"
	"chatapp/cmd/server/middlewares"
	eventlisteners "chatapp/internal/eventListeners"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/services"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/swagger"
)

type AuthHandler interface {
	Register(c *fiber.Ctx) error
	Login(c *fiber.Ctx) error
	GetProfile(c *fiber.Ctx) error
}

type ConversationHandler interface {
	CreateConversation(ctx *fiber.Ctx) error
	ShowUserTyping(ctx *fiber.Ctx) error
	ListenConversation(conn *websocket.Conn)
}

type MessageHandler interface {
	SendMessage(c *fiber.Ctx) error
	UpdateMessage(c *fiber.Ctx) error
	GetMessages(c *fiber.Ctx) error
}
type Handlers struct {
	authHandler AuthHandler
	convHandler ConversationHandler
	msgHandler  MessageHandler

	mdlwrs *middlewares.Middlewares
}

func NewHandlers(
	srvs *services.Services,
	mdlwrs *middlewares.Middlewares,
	repos *repositories.Repositories,
	evls *eventlisteners.EventListeners,
	logger logger.Logger,
) *Handlers {
	return &Handlers{
		authHandler: auth.NewHandler(logger, srvs, repos),
		convHandler: conversation.NewHandler(srvs, evls, logger),
		msgHandler:  message.NewHandler(srvs, logger),

		mdlwrs: mdlwrs,
	}
}

func (h *Handlers) RegisterRoutes(r fiber.Router) {
	r.Get("swagger/*", swagger.HandlerDefault)

	v1 := r.Group("/api/v1")

	auth := v1.Group("/auth")
	auth.Post("/signup", h.authHandler.Register)
	auth.Post("/signin", h.authHandler.Login)

	protected := v1.Group("/")
	protected.Use(h.mdlwrs.AuthMiddleware.Handle)
	protected.Use(h.mdlwrs.RateLimiterMiddleware.Handle)
	protected.Get("/profile", h.authHandler.GetProfile)

	protected.Post("/conversations", h.convHandler.CreateConversation)
	protected.Post("/conversations/:conversationId/show-user-typing", h.convHandler.ShowUserTyping)

	protected.Post("/conversations/:conversationId/messages", h.msgHandler.SendMessage)
	protected.Post("/conversations/:conversationId/messages/:messageId", h.msgHandler.UpdateMessage)
	protected.Get("/conversations/:conversationId/messages", h.msgHandler.GetMessages)

	protected.Get("/listen/conversations/:conversationId", websocket.New(h.convHandler.ListenConversation))

}
