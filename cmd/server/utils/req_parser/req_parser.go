package reqparser

import (
	"fmt"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

var v = validator.New()

func ParseReqBody(ctx *fiber.Ctx, o any) error {
	if err := ctx.BodyParser(o); err != nil {
		return fiber.ErrBadRequest
	}

	if err := v.Struct(o); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	return nil
}
