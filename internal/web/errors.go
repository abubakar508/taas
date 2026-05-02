package web

import "github.com/gofiber/fiber/v2"

type ErrorResponse struct {
	Error     string `json:"error"`
	RequestID string `json:"request_id,omitempty"`
}

func WriteError(c *fiber.Ctx, status int, message string) error {
	requestID, _ := c.Locals("request_id").(string)

	return c.Status(status).JSON(ErrorResponse{
		Error:     message,
		RequestID: requestID,
	})
}
