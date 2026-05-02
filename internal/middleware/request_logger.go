package middleware

import (
	"time"

	"github.com/abubakar508/taas/internal/logger"
	"github.com/gofiber/fiber/v2"
)

func RequestLogger(log *logger.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		requestID, _ := c.Locals("request_id").(string)
		log.Printf(
			`request_id=%s method=%s path=%s status=%d duration_ms=%d ip=%s`,
			requestID,
			c.Method(),
			c.OriginalURL(),
			c.Response().StatusCode(),
			time.Since(start).Milliseconds(),
			c.IP(),
		)

		return err
	}
}
