package message

import (
	"chatapp/cmd/server/middlewares/auth"
	reqparser "chatapp/cmd/server/utils/req_parser"
	messages_dto "chatapp/internal/dto/messages"
	chat "chatapp/internal/entities/chat"
	"chatapp/internal/logger"
	"chatapp/internal/services"
	"chatapp/internal/services/chat/utils"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	srvs   *services.Services
	logger logger.Logger
}

func NewHandler(srvs *services.Services, logger logger.Logger) *Handler {
	return &Handler{
		srvs:   srvs,
		logger: logger,
	}
}

// SendMessageRequestPayload represents the request payload for sending a message.
// swagger:model
type SendMessageRequestPayload struct {
	// Content of the message
	// required: true
	Content string `json:"content" validate:"required,min=3,max=250"`
}

// SendMessageResponse200Payload represents a successful response containing the sent message.
// swagger:model
type SendMessageResponse200Payload struct {

	// SendMessageResponse200Payload represents a successful response containing the sent message.
	// swagger:model
	Message *chat.Message `json:"message"`
}

// SendMessage sends a new message in a conversation.
// @Summary      Send a new message
// @Description  Sends a new message within a specified conversation.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        conversationId path int64 true "Conversation ID"
// @Param        payload        body      SendMessageRequestPayload true "Send Message Payload"
// @Success      200            {object}  SendMessageResponse200Payload
// @Security     UserTokenAuth
// @Router       /api/v1/conversations/{conversationId}/messages [post]
func (h *Handler) SendMessage(ctx *fiber.Ctx) error {
	reqBody := &SendMessageRequestPayload{}
	if err := reqparser.ParseReqBody(ctx, reqBody); err != nil {
		return errors.Join(err)
	}
	if ctx.Params("conversationId") == "" {
		return errors.Join(fiber.ErrBadRequest, fmt.Errorf("conversation id shoudl not be empty"))
	}
	cnvId, err := strconv.ParseInt(ctx.Params("conversationId"), 10, 64)
	if err != nil {
		h.logger.Error(ctx.Context(), fmt.Errorf("failed to parse conversation id param: %w", err), slog.Any("params", ctx.Queries()))
		return errors.Join(fiber.ErrBadRequest, err)
	}
	user := auth.MustGetUser(ctx)
	message, err := h.srvs.MessageService.SendMessage(ctx.Context(), cnvId, user.ID, reqBody.Content)
	if err != nil {
		if errors.Is(err, utils.ErrIsNotConversationParticipant) || errors.Is(err, utils.ErrConversationNotFound) {
			return errors.Join(fiber.ErrBadRequest, err)
		}

		return fiber.ErrInternalServerError
	}
	return ctx.JSON(&SendMessageResponse200Payload{
		Message: message,
	})
}

// GetMessagesResponse200Payload represents a successful response containing a list of messages.
// swagger:model
type GetMessagesResponse200Payload struct {
	// List of messages
	// required: true
	Messages []*chat.Message
}

// GetMessages retrieves messages from a conversation with optional pagination.
// @Summary      Get messages from a conversation
// @Description  Retrieves messages from a specified conversation with optional pagination parameters.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        conversationId path     int64  true "Conversation ID"
// @Param        limit          query    int    false "Number of messages to retrieve" default(10)
// @Param        lastID         query    int64  false "ID of the last message received"
// @Success      200            {object}  GetMessagesResponse200Payload
// @Security     UserTokenAuth
// @Router       /api/v1/conversations/{conversationId}/messages [get]
func (h *Handler) GetMessages(ctx *fiber.Ctx) error {
	user := auth.MustGetUser(ctx)
	cnvId, err := strconv.ParseInt(ctx.Params("conversationId"), 10, 64)
	if err != nil {
		h.logger.Error(ctx.Context(), fmt.Errorf("failed to parse query params: %w", err), slog.Any("params", ctx.Queries()))
		return fiber.ErrBadRequest
	}
	limit, err := strconv.Atoi(ctx.Query("limit", "10"))
	if err != nil {
		h.logger.Error(ctx.Context(), fmt.Errorf("failed to parse query params: %w", err), slog.Any("params", ctx.Queries()))
		return fiber.ErrBadRequest
	}
	lastIDStr := ctx.Query("lastID", "")
	var lastID *int64
	if lastIDStr != "" {
		id, err := strconv.ParseInt(lastIDStr, 10, 64)
		if err != nil {
			h.logger.Error(ctx.Context(), fmt.Errorf("failed to parse query params: %w", err), slog.Any("params", ctx.Queries()))
			return fiber.ErrBadRequest
		}
		lastID = &id
	}
	params := &messages_dto.GetMessageQueryParams{
		ConvId:         cnvId,
		Limit:          limit,
		LastReceivedId: lastID,
		UserId:         user.ID,
	}
	messages, err := h.srvs.MessageService.GetMessages(ctx.Context(), params)
	if err != nil {
		if errors.Is(err, utils.ErrIsNotConversationParticipant) || errors.Is(err, utils.ErrConversationNotFound) {
			return errors.Join(fiber.ErrBadRequest, err)
		}
		h.logger.Error(ctx.Context(), fmt.Errorf("get messages endpoint error: %w", err), slog.Any("params", params))
		return fiber.ErrInternalServerError
	}
	return ctx.JSON(&GetMessagesResponse200Payload{
		Messages: messages,
	})
}

// MessageUpdateRequestPayload represents the request payload for updating a message.
// swagger:model
type MessageUpdateRequestPayload struct {
	// Updated content of the message
	// required: true
	Content string `json:"content" validate:"required,min=3,max=250"`
}

// MessageUpdateResponse200Payload represents a successful response containing the updated message.
// swagger:model
type MessageUpdateResponse200Payload struct {
	// The updated message
	// required: true
	Message *chat.Message
}

// UpdateMessage updates the content of an existing message in a conversation.
// @Summary      Update a message
// @Description  Updates the content of a specified message within a conversation.
// @Tags         messages
// @Accept       json
// @Produce      json
// @Param        conversationId path     int64                 true  "Conversation ID"
// @Param        messageId      path     int64                 true  "Message ID"
// @Param        payload        body     MessageUpdateRequestPayload true "Update Message Payload"
// @Success      200            {object}  MessageUpdateResponse200Payload
// @Security     UserTokenAuth
// @Router       /api/v1/conversations/{conversationId}/messages/{messageId} [post]
func (h *Handler) UpdateMessage(ctx *fiber.Ctx) error {
	user := auth.MustGetUser(ctx)
	cnvId, err := strconv.ParseInt(ctx.Params("conversationId"), 10, 64)
	if err != nil {
		h.logger.Error(ctx.Context(), fmt.Errorf("failed to parse query params: %w", err), slog.String("fail_on", "conversationId"), slog.Any("params", ctx.Queries()))
		return fiber.ErrBadRequest
	}
	msgId, err := strconv.ParseInt(ctx.Params("messageId"), 10, 64)
	if err != nil {
		h.logger.Error(ctx.Context(), fmt.Errorf("failed to parse query params: %w", err), slog.String("fail_on", "messageId"), slog.Any("params", ctx.Queries()))
		return fiber.ErrBadRequest
	}
	reqBody := &MessageUpdateRequestPayload{}
	if err := reqparser.ParseReqBody(ctx, reqBody); err != nil {
		return errors.Join(fiber.ErrBadRequest, err)
	}
	message, err := h.srvs.MessageService.UpdateMessage(ctx.Context(), cnvId, msgId, user.ID, reqBody.Content)
	if err != nil {
		if errors.Is(err, utils.ErrIsNotConversationParticipant) || errors.Is(err, utils.ErrConversationNotFound) {
			return errors.Join(fiber.ErrBadRequest, err)
		}
		h.logger.Error(ctx.Context(), fmt.Errorf("update message endpoint error: %w", err))
		return fiber.ErrInternalServerError
	}
	return ctx.JSON(&MessageUpdateResponse200Payload{
		Message: message,
	})
}
