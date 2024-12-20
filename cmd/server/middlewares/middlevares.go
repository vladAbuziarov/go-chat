package middlewares

import (
	"chatapp/cmd/server/middlewares/auth"
	ratelimiter "chatapp/cmd/server/middlewares/rate_limiter"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/services"
)

type Middlewares struct {
	AuthMiddleware        *auth.Middleware
	RateLimiterMiddleware *ratelimiter.Middleware
}

func NewMiddlewares(
	srvs *services.Services,
	rps *repositories.Repositories,
	logger logger.Logger,
) *Middlewares {
	return &Middlewares{
		AuthMiddleware:        auth.NewMiddleware(srvs, rps, logger),
		RateLimiterMiddleware: ratelimiter.NewMiddleware(5, 10),
	}
}
