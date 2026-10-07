package helper

import (
	"proyek-uts/app/model"

	"github.com/gofiber/fiber/v2"
)

type WebResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

type WebResponseCursor struct {
	Success bool              `json:"success"`
	Message string            `json:"message"`
	Data    any               `json:"data"`
	Meta    *model.CursorMeta `json:"meta"`
}

func Success(c *fiber.Ctx, status int, message string, data any) error {
	return c.Status(status).JSON(WebResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

func SuccessCursor(c *fiber.Ctx, message string, data any, meta *model.CursorMeta) error {
	return c.Status(fiber.StatusOK).JSON(WebResponseCursor{
		Success: true,
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

func NoContent(c *fiber.Ctx) error {
	return c.SendStatus(fiber.StatusNoContent)
}