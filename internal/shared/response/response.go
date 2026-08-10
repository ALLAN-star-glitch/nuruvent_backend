// pkg/response/response.go

package response

import (
	"github.com/gofiber/fiber/v3"
)

type BaseResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    any `json:"data,omitempty"`
	Errors  any `json:"errors,omitempty"`
}

func Success(c fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusOK).JSON(BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func Created(c fiber.Ctx, message string, data any) error {
	return c.Status(fiber.StatusCreated).JSON(BaseResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func BadRequest(c fiber.Ctx, message string, errors any) error {
	return c.Status(fiber.StatusBadRequest).JSON(BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

func Unauthorized(c fiber.Ctx, message string, errors any) error {
	return c.Status(fiber.StatusUnauthorized).JSON(BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

func Forbidden(c fiber.Ctx, message string, errors any) error {
	return c.Status(fiber.StatusForbidden).JSON(BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

func NotFound(c fiber.Ctx, message string, errors any) error {
	return c.Status(fiber.StatusNotFound).JSON(BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

func Conflict(c fiber.Ctx, message string, errors any) error {
	return c.Status(fiber.StatusConflict).JSON(BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}

func InternalError(c fiber.Ctx, message string, errors any) error {
	return c.Status(fiber.StatusInternalServerError).JSON(BaseResponse{
		Success: false,
		Message: message,
		Errors:  errors,
	})
}