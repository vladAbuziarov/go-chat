package conversation_test

import (
	"bytes"
	"chatapp/cmd/server/handlers/chat/conversation"
	"chatapp/cmd/server/middlewares/auth"
	chatEnts "chatapp/internal/entities/chat"
	"chatapp/internal/entities/users"
	eventlisteners "chatapp/internal/eventListeners"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/repositories/fakes"
	"chatapp/internal/services"
	chat "chatapp/internal/services/chat/conversation"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupConversationApp() *fiber.App {
	app := fiber.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals(auth.UserIdContextKey{}, users.UserId(1))
		return c.Next()
	})
	fakeRepo := fakes.NewFakeConversationRepository()

	repos := repositories.Repositories{
		ConversationRepository: fakeRepo,
	}
	logger := logger.NewLogger()

	evls := eventlisteners.NewEventListeners()
	service := chat.NewService(&repos, logger, evls)
	srvs := services.Services{
		ConversationService: service,
	}

	handler := conversation.NewHandler(&srvs, evls, logger)

	app.Post("/api/v1/conversations", handler.CreateConversation)
	app.Post("/api/v1/conversations/:conversationId/show-user-typing", handler.ShowUserTyping)
	app.Get("/api/v1/listen/conversations/:conversationId", websocket.New(handler.ListenConversation))

	return app
}

func TestConversationHandler_CreateConversation(t *testing.T) {
	app := setupConversationApp()

	payload := map[string]interface{}{
		"name":            "Test Conversation",
		"is_group":        true,
		"participant_ids": []int64{2, 3},
	}
	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/conversations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var res struct {
		Conversation *chatEnts.Conversation
	}

	err = json.NewDecoder(resp.Body).Decode(&res)
	assert.NoError(t, err)
	assert.NotZero(t, res.Conversation.ID)
	assert.Equal(t, "Test Conversation", res.Conversation.Name)
	assert.True(t, res.Conversation.IsGroup)
}

func TestConversationHandler_CreateConversation_ValidationError(t *testing.T) {
	app := setupConversationApp()

	payload := map[string]interface{}{
		"name":            "T",
		"is_group":        true,
		"participant_ids": []int64{2, 3},
	}
	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/conversations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestConversationHandler_ShowUserTyping(t *testing.T) {
	app := setupConversationApp()

	payload := map[string]interface{}{
		"name":            "Test Conversation",
		"is_group":        true,
		"participant_ids": []int64{2, 3},
	}
	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/conversations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var createRes struct {
		Conversation *chatEnts.Conversation
	}
	err = json.NewDecoder(resp.Body).Decode(&createRes)
	assert.NoError(t, err)
	assert.NotZero(t, createRes.Conversation.ID)

	convIdStr := strconv.FormatInt(createRes.Conversation.ID, 10)
	typingReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/conversations/%s/show-user-typing", convIdStr), nil)
	typingReq.Header.Set("Content-Type", "application/json")
	typingResp, err := app.Test(typingReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, typingResp.StatusCode)
}

func TestConversationHandler_ShowUserTyping_ValidationError(t *testing.T) {
	app := setupConversationApp()

	payload := map[string]interface{}{
		"name":            "Test Conversation",
		"is_group":        true,
		"participant_ids": []int64{2, 3},
	}
	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/conversations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var createRes struct {
		Conversation *chatEnts.Conversation
	}
	err = json.NewDecoder(resp.Body).Decode(&createRes)
	assert.NoError(t, err)
	assert.NotZero(t, createRes.Conversation.ID)

	convIdStr := "test"
	typingReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/conversations/%s/show-user-typing", convIdStr), nil)
	typingReq.Header.Set("Content-Type", "application/json")
	typingResp, err := app.Test(typingReq)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, typingResp.StatusCode)
}

func TestConversationHandler_ShowUserTyping_AccessError(t *testing.T) {
	app := setupConversationApp()

	payload := map[string]interface{}{
		"name":            "Test Conversation",
		"is_group":        true,
		"participant_ids": []int64{2, 3},
	}
	body, err := json.Marshal(payload)
	assert.NoError(t, err)

	req := httptest.NewRequest("POST", "/api/v1/conversations", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var createRes struct {
		Conversation *chatEnts.Conversation
	}
	err = json.NewDecoder(resp.Body).Decode(&createRes)
	assert.NoError(t, err)
	assert.NotZero(t, createRes.Conversation.ID)

	convIdStr := strconv.FormatInt(20, 10)
	typingReq := httptest.NewRequest("POST", fmt.Sprintf("/api/v1/conversations/%s/show-user-typing", convIdStr), nil)
	typingReq.Header.Set("Content-Type", "application/json")
	typingResp, err := app.Test(typingReq)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, typingResp.StatusCode)
}
