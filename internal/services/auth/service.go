package user

import (
	user_dto "chatapp/internal/dto/user"
	"chatapp/internal/entities/users"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/services/hash"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var (
	ErrCannotCreateUser = errors.New("cannot create user")
	ErrUnAutorize       = errors.New("cannot authorize user")
)

type Service struct {
	logger     logger.Logger
	hasService *hash.Service
	repos      *repositories.Repositories
}

func NewService(logger logger.Logger, repos *repositories.Repositories) *Service {
	return &Service{
		logger:     logger,
		hasService: hash.NewService(),
		repos:      repos,
	}
}

func (s *Service) Register(ctx context.Context, user *user_dto.CreateUserDTO) (*users.User, error) {
	s.logger.Info(ctx, "reuest for new user creation", slog.Any("user", user))

	hash, err := s.hasService.HashPassword(user.Password)
	if err != nil {
		return nil, errors.Join(ErrCannotCreateUser, err)
	}
	user.Password = hash

	newUser, err := s.repos.UserRepository.Create(ctx, user)
	if err != nil {
		s.logger.Error(ctx, errors.Join(ErrCannotCreateUser, err), slog.Any("user", user))
		return nil, errors.Join(ErrCannotCreateUser, err)
	}

	s.logger.Info(ctx, "created new user", slog.Any("user", newUser))
	return newUser, nil
}

func (s *Service) Login(ctx context.Context, email, password string) (*users.User, error) {
	user, err := s.repos.UserRepository.GetUserByEmail(ctx, email)
	if err != nil {
		s.logger.Info(ctx, "Cannot authorize user due to error",
			slog.Any("error", err),
			slog.String("email", email),
			slog.String("password", password))
		return nil, err
	}
	if user == nil {
		s.logger.Info(ctx, fmt.Sprintf("cannot found user with email: %s", email))
		return nil, ErrUnAutorize
	}
	passOk, err := s.hasService.CompareHashWithPassword(user.Password, password)
	if err != nil {
		s.logger.Info(ctx, "Cannot authorize user due to error",
			slog.Any("error", err),
			slog.String("email", email),
			slog.String("password", password))
		return nil, err
	}
	if !passOk {
		return nil, errors.Join(ErrUnAutorize, fmt.Errorf("password hash check failed"))
	}

	return user, nil
}
