package conversation

import (
	"chatapp/cmd/server/middlewares/auth"
	reqparser "chatapp/cmd/server/utils/req_parser"
	chat "chatapp/internal/entities/chat"
	eventlisteners "chatapp/internal/eventListeners"
	"chatapp/internal/logger"
	"chatapp/internal/services"
	"chatapp/internal/services/chat/utils"
	"chatapp/internal/services/events"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	srvs   *services.Services
	logger logger.Logger
	evls   *eventlisteners.EventListeners
}

func NewHandler(srvs *services.Services, evls *eventlisteners.EventListeners, logger logger.Logger) *Handler {
	return &Handler{
		srvs:   srvs,
		evls:   evls,
		logger: logger,
	}
}

// CreateConversationRequestPayload represents the request payload for creating a conversation.
// swagger:model
type CreateConversationRequestPayload struct {
	// Name of the conversation
	// required: true
	Name string `json:"name" validate:"required,min=3,max=50"`
	// Indicates if the conversation is a group chat
	// required: true
	IsGroup bool `json:"is_group" validate:"required,boolean"`
	// IDs of participants to be added to the conversation
	// required: true
	ParticipantIDs []int64 `json:"participant_ids" validate:"required"`
}
// CreateConversationResponse200Payload represents a successful response containing the created conversation.
// swagger:model
type CreateConversationResponse200Payload struct {
	Conversation *chat.Conversation `json:"conversation"`
}

// @Summary      Create a new conversation
// @Description  Creates a new conversation with specified participants.
// @Tags         conversations
// @Accept       json
// @Produce      json
// @Param        payload  body      CreateConversationRequestPayload  true  "Create Conversation Request"
// @Success      200      {object}  CreateConversationResponse200Payload
// @Router       /conversations [post]
// @Security UserTokenAuth
func (h *Handler) CreateConversation(ctx *fiber.Ctx) error {
	reqBody := &CreateConversationRequestPayload{}
	if err := reqparser.ParseReqBody(ctx, reqBody); err != nil {
		return errors.Join(fiber.ErrBadRequest, err)
	}
	u := auth.MustGetUser(ctx)

	if !slices.Contains(reqBody.ParticipantIDs, u.ID) {
		reqBody.ParticipantIDs = append(reqBody.ParticipantIDs, u.ID)
	}

	conv, err := h.srvs.ConversationService.CreateConversation(ctx.Context(), reqBody.Name, reqBody.IsGroup, reqBody.ParticipantIDs)
	if err != nil {
		return errors.Join(fiber.ErrInternalServerError, err)
	}

	return ctx.JSON(&CreateConversationResponse200Payload{
		Conversation: conv,
	})
}

// listen ws with conversation messages
func (h *Handler) ListenConversation(conn *websocket.Conn) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	defer conn.Close()

	go func() {
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				cancel()
				break
			}
		}
	}()
	if conn.Params("conversationId") == "" {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"invalid conversation ID"}`))
		return
	}
	convId, err := strconv.ParseInt(conn.Params("conversationId"), 10, 64)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(`{"error":"failed to parse conversation ID"}`))
		return
	}

	ch := make(chan events.Event)
	h.evls.ChatEventListener.SubscribeChannel(convId, ch)
	defer h.evls.ChatEventListener.UnsubscribeChannel(convId, ch)

	for msg := range ch {
		if err := conn.WriteJSON(msg); err != nil {
			h.logger.Error(ctx, fmt.Errorf("cannot write websocket message: %w", err))
			return
		}
	}

}

// ShowUserTypingResponse200Payload represents a successful response for showing user typing status.
// swagger:model
type ShowUserTypingResponse200Payload struct {
	// Indicates whether the operation was successful.
    // required: true
	Success bool `json:"success"`
}

// @Summary      Show user typing status
// @Description  Records that a user is typing in a conversation.
// @Tags         conversations
// @Param        conversationId path int64 true "Conversation ID"
// @Produce      json
// @Success      200      {object}  ShowUserTypingResponse200Payload
// @Security     UserTokenAuth
// @Router       /api/v1/conversations/{conversationId}/show-user-typing [post]
func (h *Handler) ShowUserTyping(ctx *fiber.Ctx) error {
	user := auth.MustGetUser(ctx)

	cnvId, err := strconv.ParseInt(ctx.Params("conversationId"), 10, 64)
	if err != nil {
		h.logger.Error(ctx.Context(), fmt.Errorf("failed to parse query params: %w", err), slog.Any("params", ctx.AllParams()))
		return fiber.ErrBadRequest
	}
	if err := h.srvs.ConversationService.PostUserTyping(ctx.Context(), user.ID, cnvId); err != nil {
		if errors.Is(err, utils.ErrIsNotConversationParticipant) || errors.Is(err, utils.ErrConversationNotFound) {
			return errors.Join(fiber.ErrBadRequest, err)
		}
		h.logger.Error(ctx.Context(), fmt.Errorf("show user typing endpoint error: %w", err))
		return fiber.ErrInternalServerError
	}
	return ctx.JSON(&ShowUserTypingResponse200Payload{
		Success: true,
	})
}
