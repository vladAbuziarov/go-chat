package auth

import (
	fu "chatapp/cmd/server/utils/fiber"
	"chatapp/internal/entities/users"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/services"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

var ErrUserIdExpected = errors.New("user ID expected")

type Middleware struct {
	srvs   *services.Services
	repos  *repositories.Repositories
	logger logger.Logger
}

func NewMiddleware(srvs *services.Services, repos *repositories.Repositories, logger logger.Logger) *Middleware {
	return &Middleware{
		srvs:   srvs,
		repos:  repos,
		logger: logger,
	}
}

const authHeaderName = "X-User-Token"

type UserIdContextKey struct{}

func (m *Middleware) Handle(ctx *fiber.Ctx) error {
	token := ctx.Get(authHeaderName)
	if token == "" {
		return fiber.ErrUnauthorized
	}
	userId, err := m.srvs.JwtService.VerifyAuthToken(token)
	if err != nil {
		return fiber.ErrUnauthorized
	}
	user, err := m.repos.UserRepository.GetUserById(ctx.Context(), users.UserId(*userId))
	if err != nil {
		return fmt.Errorf("failed to get user by id: %w", err)
	}

	if user == nil {
		return fiber.ErrNotFound
	}
	ctx.Locals(UserIdContextKey{}, user.ID)
	fu.SetLoggerAttrs(ctx, slog.Int64("user_id", *userId))
	return ctx.Next()
}

func mustBeUser(v any) users.UserId {
	if u, ok := v.(users.UserId); ok {
		return u
	}

	panic(ErrUserIdExpected)
}

func MustGetUser(ctx *fiber.Ctx) users.UserId {
	val := ctx.Locals(UserIdContextKey{}).(users.UserId)
	return mustBeUser(val)
}
