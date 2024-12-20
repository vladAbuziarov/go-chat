package auth

import (
	"chatapp/cmd/server/middlewares/auth"
	reqparser "chatapp/cmd/server/utils/req_parser"
	user_dto "chatapp/internal/dto/user"
	"chatapp/internal/entities/users"
	"chatapp/internal/logger"
	"chatapp/internal/repositories"
	"chatapp/internal/services"
	authServ "chatapp/internal/services/auth"
	"errors"
	"fmt"
	"log/slog"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	srvs   *services.Services
	repos  *repositories.Repositories
	logger logger.Logger
}

func NewHandler(logger logger.Logger, srvs *services.Services, repos *repositories.Repositories) *Handler {
	return &Handler{
		srvs:   srvs,
		repos:  repos,
		logger: logger,
	}
}

// register endpoint
// RegisterRequestPayload represents the request payload for user registration.
// swagger:model
type RegisterRequestPayload struct {
	// Username of the user
	// required: true
	UserName string `json:"name" validate:"required,min=3,max=20,alphanum"`
	// Password of the user
	// required: true
	Password string `json:"password" validate:"required,min=8,max=20,alphanum"`
	// Email address of the user
	// required: true
	Email string `json:"email" validate:"required,email"`
}

// RegisterResponse200Payload represents a successful response containing the registered user.
// swagger:model
type RegisterResponse200Payload struct {
	// The registered user
	// required: true
	User *users.User `json:"user"`
}

// @Summary      Register a new user
// @Description  Registers a new user with the provided details.
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        payload  body      RegisterRequestPayload  true  "Register Request Payload"
// @Success      200      {object}  RegisterResponse200Payload
// @Router       /api/v1/auth/signup [post]
func (h *Handler) Register(ctx *fiber.Ctx) error {
	reqBody := &RegisterRequestPayload{}

	if err := reqparser.ParseReqBody(ctx, reqBody); err != nil {
		return errors.Join(fiber.ErrBadRequest, err)
	}

	user, err := h.srvs.UserService.Register(ctx.Context(), &user_dto.CreateUserDTO{
		Name:     reqBody.UserName,
		Email:    reqBody.Email,
		Password: reqBody.Password,
	})
	if err != nil {
		return errors.Join(fiber.ErrInternalServerError, err)
	}
	return ctx.JSON(&RegisterResponse200Payload{
		User: user,
	})
}

// login endpoint
// LoginRequestPayload represents the request payload for user login.
// swagger:model
type LoginRequestPayload struct {
	// Password of the user
	// required: true
	Password string `json:"password" validate:"required,min=8,max=20,alphanum"`
	// Email address of the user
	// required: true
	Email string `json:"email" validate:"required,email"`
}

// LoginResp200Body represents a successful login response containing a JWT token.
// swagger:model
type LoginResp200Body struct {
	// JWT token for authenticated requests
	// required: true
	Token string `json:"token"`
}

// @Summary      User login
// @Description  Authenticates a user and returns a JWT token.
// @Tags         authentication
// @Accept       json
// @Produce      json
// @Param        payload  body      LoginRequestPayload  true  "Login Request Payload"
// @Success      200      {object}  LoginResp200Body
// @Router       /api/v1/auth/signin [post]
func (h *Handler) Login(ctx *fiber.Ctx) error {
	reqBody := &LoginRequestPayload{}
	if err := reqparser.ParseReqBody(ctx, reqBody); err != nil {
		return errors.Join(fiber.ErrBadRequest, err)
	}
	user, err := h.srvs.UserService.Login(ctx.Context(), reqBody.Email, reqBody.Password)
	if err != nil {
		if errors.Is(err, authServ.ErrUnAutorize) {
			return fiber.ErrForbidden
		}
		return errors.Join(fiber.ErrInternalServerError, err)
	}
	token, err := h.srvs.JwtService.CreateToken(ctx.Context(), user.ID)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	resp := &LoginResp200Body{
		Token: token,
	}
	return ctx.JSON(resp)
}

// profile endpoint
// ProfileResp200Body represents a successful response containing user profile.
// swagger:model
type ProfileResp200Body struct {
	// The user's profile information
	// required: true
	User *users.User `json:"profile"`
}

// @Summary      Get user profile
// @Description  Retrieves the authenticated user's profile.
// @Tags         authentication
// @Produce      json
// @Success      200      {object}  ProfileResp200Body
// @Security     UserTokenAuth
// @Router       /api/v1/auth/profile [get]
func (h *Handler) GetProfile(ctx *fiber.Ctx) error {
	u := auth.MustGetUser(ctx)

	user, err := h.repos.UserRepository.GetUserById(ctx.Context(), u.ID)
	if err != nil {
		h.logger.Error(ctx.Context(), fmt.Errorf("user fetch error: %w", err), slog.Int64("userId", u.ID))
		return fiber.ErrInternalServerError
	}
	if user == nil {
		return fiber.ErrUnauthorized
	}
	return ctx.JSON(ProfileResp200Body{
		User: user,
	})
}
