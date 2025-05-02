package user

import (
	user_dto "chatapp/internal/dto/user"
	"chatapp/internal/entities/users"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"context"
	"errors"
	"testing"

	repoMocks "chatapp/internal/repositories/mocks"
	authMocks "chatapp/internal/services/auth/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
)

func TestService_Register_Success(t *testing.T) {
	// Create a new GoMock controller.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks using GoMock.
	mockHashService := authMocks.NewMockHashService(ctrl)
	mockUserRepository := repoMocks.NewMockUserRepositoryInterface(ctrl)

	// Use the generated mock for the repository in the Repositories struct.
	repos := &repositories.Repositories{
		// <user__selection>userRepo</user__selection> replaced by the GoMock mock:
		UserRepository: mockUserRepository,
	}

	// Initialize the service with the mocks.
	svc := NewService(logger.NewLogger(), mockHashService, repos)

	ctx := context.Background()
	createDTO := &user_dto.CreateUserDTO{
		Name:     "testuser",         // valid (>=3 characters)
		Email:    "test@example.com", // non-empty email
		Password: "password123",      // valid (>=8 characters)
	}
	hashedPassword := "hashedpassword123"

	// Set up expectations:
	// 1. Expect the hash service to hash the password.
	mockHashService.
		EXPECT().
		HashPassword(createDTO.Password).
		Return(hashedPassword, nil)

	// 2. Expect the repository to be called to create the user.
	expectedUser := &users.User{
		ID:       1,
		Username: createDTO.Name,
		Email:    createDTO.Email,
		Password: hashedPassword,
	}
	mockUserRepository.
		EXPECT().
		Create(ctx, createDTO).
		Return(expectedUser, nil)

	// Act: Call the Register method.
	user, err := svc.Register(ctx, createDTO)

	// Assert the results.
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, expectedUser.ID, user.ID)
	assert.Equal(t, hashedPassword, user.Password)
}

func TestService_Register_Hash_Password_Error(t *testing.T) {
	// Create a new GoMock controller.
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create mocks using GoMock.
	mockHashService := authMocks.NewMockHashService(ctrl)
	mockUserRepository := repoMocks.NewMockUserRepositoryInterface(ctrl)

	// Use the generated mock for the repository in the Repositories struct.
	repos := &repositories.Repositories{
		// <user__selection>userRepo</user__selection> replaced by the GoMock mock:
		UserRepository: mockUserRepository,
	}

	// Initialize the service with the mocks.
	svc := NewService(logger.NewLogger(), mockHashService, repos)

	ctx := context.Background()
	createDTO := &user_dto.CreateUserDTO{
		Name:     "testuser",         // valid (>=3 characters)
		Email:    "test@example.com", // non-empty email
		Password: "qwkemlqwkmelqwke", // valid (>=8 characters)
	}

	// Set up expectations:
	// 1. Expect the hash service to hash the password.
	mockHashService.
		EXPECT().
		HashPassword(createDTO.Password).
		Return("", errors.New("password hashing error"))

	// Act: Call the Register method.
	user, err := svc.Register(ctx, createDTO)

	// Assert the results.
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, ErrCannotCreateUser)
}

func TestService_Register_Validation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create dummy mocks.
	mockHashService := authMocks.NewMockHashService(ctrl)
	mockUserRepository := repoMocks.NewMockUserRepositoryInterface(ctrl)
	repos := &repositories.Repositories{
		UserRepository: mockUserRepository,
	}
	// Initialize the service.
	svc := NewService(logger.NewLogger(), mockHashService, repos)
	ctx := context.Background()

	tests := []struct {
		name                 string
		dto                  *user_dto.CreateUserDTO
		expectedErrSubstring string
	}{
		{
			name: "invalid name",
			dto: &user_dto.CreateUserDTO{
				Name:     "ab", // too short
				Email:    "valid@example.com",
				Password: "validpassword",
			},
			expectedErrSubstring: ErrNameValidation.Error(),
		},
		{
			name: "empty email",
			dto: &user_dto.CreateUserDTO{
				Name:     "validname",
				Email:    "",
				Password: "validpassword",
			},
			expectedErrSubstring: ErrEmailValidation.Error(),
		},
		{
			name: "short password",
			dto: &user_dto.CreateUserDTO{
				Name:     "validname",
				Email:    "valid@example.com",
				Password: "short", // too short (expected: >= 8 characters)
			},
			expectedErrSubstring: ErrPasswordValidation.Error(),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			user, err := svc.Register(ctx, tc.dto)
			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tc.expectedErrSubstring)
		})
	}
}

func TestService_Login_Success(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHashService := authMocks.NewMockHashService(ctrl)
	mockUserRepository := repoMocks.NewMockUserRepositoryInterface(ctrl)
	repos := &repositories.Repositories{
		UserRepository: mockUserRepository,
	}
	svc := NewService(logger.NewLogger(), mockHashService, repos)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	storedHashedPassword := "hashedpassword123"
	existingUser := &users.User{
		ID:       1,
		Username: "testuser",
		Email:    email,
		Password: storedHashedPassword,
	}

	mockUserRepository.
		EXPECT().
		GetUserByEmail(ctx, email).
		Return(existingUser, nil)
	mockHashService.
		EXPECT().
		CompareHashWithPassword(storedHashedPassword, password).
		Return(true, nil)

	user, err := svc.Login(ctx, email, password)
	assert.NoError(t, err)
	assert.NotNil(t, user)
	assert.Equal(t, existingUser.ID, user.ID)
}

func TestService_Login_Validation(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// Create dummy mocks.
	mockHashService := authMocks.NewMockHashService(ctrl)
	mockUserRepository := repoMocks.NewMockUserRepositoryInterface(ctrl)
	repos := &repositories.Repositories{
		UserRepository: mockUserRepository,
	}
	svc := NewService(logger.NewLogger(), mockHashService, repos)
	ctx := context.Background()

	loginTests := []struct {
		name                 string
		email                string
		password             string
		expectedErrSubstring string
	}{
		{
			name:                 "empty email",
			email:                "",
			password:             "password123",
			expectedErrSubstring: "email validation error",
		},
		{
			name:                 "short password",
			email:                "test@example.com",
			password:             "short", // too short
			expectedErrSubstring: "password validation error",
		},
	}

	for _, tc := range loginTests {
		t.Run(tc.name, func(t *testing.T) {
			user, err := svc.Login(ctx, tc.email, tc.password)
			assert.Error(t, err)
			assert.Nil(t, user)
			assert.Contains(t, err.Error(), tc.expectedErrSubstring)
		})
	}
}

func TestService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHashService := authMocks.NewMockHashService(ctrl)
	mockUserRepository := repoMocks.NewMockUserRepositoryInterface(ctrl)
	repos := &repositories.Repositories{
		UserRepository: mockUserRepository,
	}
	svc := NewService(logger.NewLogger(), mockHashService, repos)

	ctx := context.Background()
	email := "nonexistent@example.com"
	password := "password123"

	mockUserRepository.
		EXPECT().
		GetUserByEmail(ctx, email).
		Return(nil, nil)

	user, err := svc.Login(ctx, email, password)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, ErrUnAutorize, err)
}
func TestService_Login_CompareHashError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHashService := authMocks.NewMockHashService(ctrl)
	mockUserRepository := repoMocks.NewMockUserRepositoryInterface(ctrl)
	repos := &repositories.Repositories{
		UserRepository: mockUserRepository,
	}
	svc := NewService(logger.NewLogger(), mockHashService, repos)

	ctx := context.Background()
	email := "test@example.com"
	password := "password123"
	storedHashedPassword := "hashedpassword123"
	existingUser := &users.User{
		ID:       1,
		Username: "testuser",
		Email:    email,
		Password: storedHashedPassword,
	}

	mockUserRepository.
		EXPECT().
		GetUserByEmail(ctx, email).
		Return(existingUser, nil)
	compareError := errors.New("compare error")
	mockHashService.
		EXPECT().
		CompareHashWithPassword(storedHashedPassword, password).
		Return(false, compareError)

	user, err := svc.Login(ctx, email, password)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Equal(t, compareError, err)
}
func TestService_Login_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockHashService := authMocks.NewMockHashService(ctrl)
	mockUserRepository := repoMocks.NewMockUserRepositoryInterface(ctrl)
	repos := &repositories.Repositories{
		UserRepository: mockUserRepository,
	}
	svc := NewService(logger.NewLogger(), mockHashService, repos)

	ctx := context.Background()
	email := "test@example.com"
	password := "wrongpassword"
	storedHashedPassword := "hashedpassword123"
	existingUser := &users.User{
		ID:       1,
		Username: "testuser",
		Email:    email,
		Password: storedHashedPassword,
	}

	mockUserRepository.
		EXPECT().
		GetUserByEmail(ctx, email).
		Return(existingUser, nil)
	mockHashService.
		EXPECT().
		CompareHashWithPassword(storedHashedPassword, password).
		Return(false, nil)

	user, err := svc.Login(ctx, email, password)
	assert.Error(t, err)
	assert.Nil(t, user)
	assert.Contains(t, err.Error(), "cannot authorize user")
	assert.Contains(t, err.Error(), "password hash check failed")
}
