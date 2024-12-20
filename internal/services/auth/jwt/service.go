package jwt

import (
	"chatapp/internal/config"
	"chatapp/internal/logger"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const tokenTTL = 5 * time.Hour

type Service struct {
	cfg    *config.Config
	logger logger.Logger
}

func NewService(cfg *config.Config, logger logger.Logger) *Service {
	return &Service{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Service) CreateToken(ctx context.Context, userId int64) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   strconv.FormatInt(userId, 10),
		IssuedAt:  jwt.NewNumericDate(now),
		ExpiresAt: jwt.NewNumericDate(now.Add(tokenTTL)),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signedToken, err := token.SignedString([]byte(s.cfg.JWTSecret))
	if err != nil {
		s.logger.Error(ctx, fmt.Errorf("failed to sign token: %w", err), slog.Int64("userId", userId))
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signedToken, nil
}
func (s *Service) VerifyAuthToken(token string) (*int64, error) {
	claims := jwt.RegisteredClaims{}

	_, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	userId, err := strconv.ParseInt(claims.Subject, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("fail to parse subejct to int64: %w", err)
	}
	return &userId, nil
}
