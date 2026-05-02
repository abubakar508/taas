package vodacom

import (
	"context"
	"time"

	vodatransaction "github.com/abubakar508/taas/internal/providers/drc/vodacom/transaction"
	"github.com/abubakar508/taas/internal/web"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TransactionHandler struct {
	service *vodatransaction.Service
}

func NewTransactionHandler(pool *pgxpool.Pool) *TransactionHandler {
	return &TransactionHandler{
		service: vodatransaction.NewService(pool),
	}
}

func (h *TransactionHandler) Register(r fiber.Router) {
	r.Get("/transactions/:id", h.get)
}

func (h *TransactionHandler) get(c *fiber.Ctx) error {
	partner, err := web.GetPartnerFromCtx(c)
	if err != nil {
		return web.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	txID := c.Params("id")
	if txID == "" {
		return web.WriteError(c, fiber.StatusBadRequest, "transaction id is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := h.service.Get(ctx, partner.ID, txID)
	if err != nil {
		return web.WriteError(c, fiber.StatusNotFound, "transaction not found")
	}

	return c.JSON(resp)
}
