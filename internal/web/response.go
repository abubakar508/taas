package web

import "github.com/gofiber/fiber/v2"

func WriteJSON(c *fiber.Ctx, status int, data interface{}) error {
	return c.Status(status).JSON(data)
}
