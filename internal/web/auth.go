package web

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/abubakar508/taas/internal/domain/partners"
	"github.com/gofiber/fiber/v2"
)

type AuthMiddleware struct {
	repo partners.Repository
}

func NewAuthMiddleware(repo partners.Repository) *AuthMiddleware {
	return &AuthMiddleware{repo: repo}
}

func (m *AuthMiddleware) RequirePartnerAuth() fiber.Handler {
	return func(c *fiber.Ctx) error {
		appName := c.Get("X-App-Name")
		apiKey := c.Get("X-Api-Key")

		if appName == "" || apiKey == "" {
			return WriteError(c, fiber.StatusUnauthorized, "missing authentication headers")
		}

		hashed := sha256.Sum256([]byte(apiKey))
		keyHash := hex.EncodeToString(hashed[:])

		partner, err := m.repo.FindActiveByAppNameAndKeyHash(c.Context(), appName, keyHash)
		if err != nil {
			return WriteError(c, fiber.StatusUnauthorized, "invalid credentials")
		}

		c.Locals("partner", partner)
		return c.Next()
	}
}

func GetPartnerFromCtx(c *fiber.Ctx) (*partners.Partner, error) {
	p, ok := c.Locals("partner").(*partners.Partner)
	if !ok || p == nil {
		return nil, fiber.ErrUnauthorized
	}
	return p, nil
}
