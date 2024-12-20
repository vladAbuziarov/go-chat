package ratelimiter

import (
	"chatapp/cmd/server/middlewares/auth"
	"fmt"
	"sync"

	"github.com/gofiber/fiber/v2"
	"golang.org/x/time/rate"
)

type Middleware struct {
	limiters map[string]*rate.Limiter
	mu       sync.Mutex
	r        rate.Limit
	b        int
}

func NewMiddleware(r rate.Limit, b int) *Middleware {
	return &Middleware{
		limiters: make(map[string]*rate.Limiter),
		r:        r,
		b:        b,
	}
}

func (rl *Middleware) getLimiter(key string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	limiter, exists := rl.limiters[key]
	if !exists {
		limiter = rate.NewLimiter(rl.r, rl.b)
		rl.limiters[key] = limiter
	}
	return limiter
}

func (rl *Middleware) Handle(c *fiber.Ctx) error {
	userID := auth.MustGetUser(c).ID
	key := fmt.Sprintf("user-%d", userID)
	limiter := rl.getLimiter(key)

	if !limiter.Allow() {
		return fiber.ErrTooManyRequests
	}

	return c.Next()
}
