package auth_test

import (
	"bytes"
	"chatapp/cmd/server/handlers"
	"chatapp/cmd/server/middlewares"
	"chatapp/cmd/server/middlewares/auth"
	ratelimiter "chatapp/cmd/server/middlewares/rate_limiter"
	"chatapp/internal/config"
	"chatapp/internal/entities/users"
	eventlisteners "chatapp/internal/eventListeners"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/repositories/fakes"
	"chatapp/internal/services"
	user "chatapp/internal/services/auth"
	"chatapp/internal/services/auth/jwt"
	"chatapp/internal/services/hash"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func setupApp() *fiber.App {
	fakeRepo := fakes.NewFakeUserRepository()
	reposWrapper := &repositories.Repositories{
		UserRepository: fakeRepo,
	}
	logger := logger.NewLogger()
	hashSvc := hash.NewService()

	authService := user.NewService(logger, hashSvc, reposWrapper)

	servs := &services.Services{
		UserService: authService,
		JwtService: jwt.NewService(5*time.Hour, &config.Config{
			JWTSecret: "test",
		}, logger),
	}

	serverHandlers := handlers.NewHandlers(servs, &middlewares.Middlewares{
		AuthMiddleware:        auth.NewMiddleware(servs, reposWrapper, logger),
		RateLimiterMiddleware: ratelimiter.NewMiddleware(5, 10),
	}, reposWrapper, eventlisteners.NewEventListeners(), logger)

	app := fiber.New()
	serverHandlers.RegisterRoutes(app)
	return app
}
func TestAuthHandler_RegisterAndLogin(t *testing.T) {
	app := setupApp()

	regPayload := map[string]string{
		"name":     "integrationuser",
		"email":    "integration@example.com",
		"password": "integrationpass",
	}
	regBody, err := json.Marshal(regPayload)
	assert.NoError(t, err)

	regReq := httptest.NewRequest("POST", "/api/v1/auth/signup", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, regResp.StatusCode)

	var regResponse struct {
		User users.User
	}
	err = json.NewDecoder(regResp.Body).Decode(&regResponse)
	assert.NoError(t, err)
	assert.NotZero(t, regResponse.User.ID)
	assert.Equal(t, regPayload["name"], regResponse.User.Username)
	assert.Equal(t, regPayload["email"], regResponse.User.Email)

	loginPayload := map[string]string{
		"email":    regPayload["email"],
		"password": regPayload["password"],
	}
	loginBody, err := json.Marshal(loginPayload)
	assert.NoError(t, err)

	loginReq := httptest.NewRequest("POST", "/api/v1/auth/signin", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginResponse struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResponse.Token)

	badLoginPayload := map[string]string{
		"email":    regPayload["email"],
		"password": "wrongpassword",
	}
	badLoginBody, err := json.Marshal(badLoginPayload)
	assert.NoError(t, err)

	badLoginReq := httptest.NewRequest("POST", "/api/v1/auth/signin", bytes.NewReader(badLoginBody))
	badLoginReq.Header.Set("Content-Type", "application/json")
	badLoginResp, err := app.Test(badLoginReq)
	assert.NoError(t, err)

	assert.NotEqual(t, http.StatusOK, badLoginResp.StatusCode)
}

func TestAuthHandlerIntegration_GetProfile(t *testing.T) {
	app := setupApp()

	// Register a new user.
	regPayload := map[string]string{
		"name":     "profileuser",
		"email":    "profile@example.com",
		"password": "profilepass",
	}
	regBody, err := json.Marshal(regPayload)
	assert.NoError(t, err)

	regReq := httptest.NewRequest("POST", "/api/v1/auth/signup", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, regResp.StatusCode)

	// Login the user.
	loginPayload := map[string]string{
		"email":    regPayload["email"],
		"password": regPayload["password"],
	}
	loginBody, err := json.Marshal(loginPayload)
	assert.NoError(t, err)

	loginReq := httptest.NewRequest("POST", "/api/v1/auth/signin", bytes.NewReader(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")
	loginResp, err := app.Test(loginReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginResponse struct {
		Token string `json:"token"`
	}
	err = json.NewDecoder(loginResp.Body).Decode(&loginResponse)
	assert.NoError(t, err)
	assert.NotEmpty(t, loginResponse.Token)

	//get user proifle
	profileReq := httptest.NewRequest("GET", "/api/v1/profile", nil)
	profileReq.Header.Set("X-User-Token", loginResponse.Token)
	profileResp, err := app.Test(profileReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, profileResp.StatusCode)

	var profileResponse struct {
		User *users.User `json:"profile"`
	}
	err = json.NewDecoder(profileResp.Body).Decode(&profileResponse)
	assert.NoError(t, err)
	assert.NotZero(t, profileResponse.User.ID)
	assert.Equal(t, regPayload["name"], profileResponse.User.Username)
	assert.Equal(t, regPayload["email"], profileResponse.User.Email)
}

func TestAuthHandler_RegisterValidationError(t *testing.T) {
	app := setupApp()

	regPayload := map[string]string{
		"name":     "integrationuser",
		"email":    "qw", //invalid email
		"password": "integrationpass",
	}
	regBody, err := json.Marshal(regPayload)
	assert.NoError(t, err)

	regReq := httptest.NewRequest("POST", "/api/v1/auth/signup", bytes.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regResp, err := app.Test(regReq)

	assert.NoError(t, err)
	assert.Equal(t, http.StatusBadRequest, regResp.StatusCode)
}
