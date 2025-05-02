package user

import (
	user_dto "chatapp/internal/dto/user"
	"chatapp/internal/entities/users"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"context"
	"errors"
	"fmt"
	"log/slog"
)

var (
	ErrCannotCreateUser   = errors.New("cannot create user")
	ErrUnAutorize         = errors.New("cannot authorize user")
	ErrEmailValidation    = errors.New("email should not be empty")
	ErrNameValidation     = errors.New("user name should not be empty and must contain more than 3 symbols")
	ErrPasswordValidation = fmt.Errorf("password should not be empty and must contain more than %d symbols", PASSWORD_MIN_LENGTH)
)

const PASSWORD_MIN_LENGTH = 8

type HashService interface {
	HashPassword(password string) (string, error)
	CompareHashWithPassword(hash, password string) (bool, error)
}
type Service struct {
	logger      logger.Logger
	hashService HashService
	repos       *repositories.Repositories
}

func NewService(logger logger.Logger, hashService HashService, repos *repositories.Repositories) *Service {
	return &Service{
		logger:      logger,
		repos:       repos,
		hashService: hashService,
	}
}

func (s *Service) Register(ctx context.Context, user *user_dto.CreateUserDTO) (*users.User, error) {
	s.logger.Info(ctx, "reuest for new user creation", slog.Any("user", user))

	if err := s.validateRegisterUserBody(user); err != nil {
		s.logger.Error(ctx, errors.Join(ErrCannotCreateUser, err), slog.Any("request_body", user))
		return nil, errors.Join(ErrCannotCreateUser, err)
	}
	hash, err := s.hashService.HashPassword(user.Password)
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
	if errEmail := validateEmail(email); errEmail != nil {
		return nil, fmt.Errorf("email validation error: %v", errEmail)
	}
	if errPswd := validatePassword(password); errPswd != nil {
		return nil, fmt.Errorf("password validation error: %v", errPswd)
	}
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
	passOk, err := s.hashService.CompareHashWithPassword(user.Password, password)
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

func (s *Service) validateRegisterUserBody(user *user_dto.CreateUserDTO) error {
	if len(user.Name) < 3 {
		return ErrNameValidation
	}
	if err := validateEmail(user.Email); err != nil {
		return err
	}
	if err := validatePassword(user.Password); err != nil {
		return err
	}
	return nil
}
func validateEmail(email string) error {
	if len(email) == 0 {
		return ErrEmailValidation
	}
	return nil
}
func validatePassword(password string) error {
	if len(password) < PASSWORD_MIN_LENGTH {
		return ErrPasswordValidation
	}
	return nil
}
