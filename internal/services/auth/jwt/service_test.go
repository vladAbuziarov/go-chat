package jwt

import (
	"chatapp/internal/config"
	"chatapp/internal/entities/users"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTService_CreateAndVerifyToken(t *testing.T) {
	ctx := context.Background()
	// Create a dummy config with a known secret.
	dummyCfg := &config.Config{
		JWTSecret: "mysecretkey",
	}
	svc := NewService(5*time.Hour, dummyCfg, nil) // Passing nil for logger in tests.

	// Create a token for a user with ID 42.
	userID := users.UserId(42)
	tokenString, err := svc.CreateToken(ctx, userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)

	// Verify the token.
	verifiedUserID, err := svc.VerifyAuthToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, verifiedUserID)
	assert.Equal(t, int64(userID), *verifiedUserID)
}

func TestJWTService_VerifyInvalidToken(t *testing.T) {
	dummyCfg := &config.Config{
		JWTSecret: "mysecretkey",
	}
	svc := NewService(5*time.Hour, dummyCfg, nil)

	// Test with an invalid token string.
	invalidToken := "invalid.token.string"
	userID, err := svc.VerifyAuthToken(invalidToken)
	assert.Error(t, err)
	assert.Nil(t, userID)
}

func TestJWTService_TokenExpiry(t *testing.T) {
	ctx := context.Background()
	dummyCfg := &config.Config{
		JWTSecret: "mysecretkey",
	}
	svc := NewService(time.Second*1, dummyCfg, nil)

	// Temporarily override tokenTTL to a short duration.

	userID := users.UserId(100)
	tokenString, err := svc.CreateToken(ctx, userID)
	assert.NoError(t, err)
	time.Sleep(2 * time.Second) // Wait for token to expire.

	verifiedUserID, err := svc.VerifyAuthToken(tokenString)
	assert.Error(t, err)
	assert.Nil(t, verifiedUserID)
}
