package vodacom

import (
	"context"
	"time"

	"github.com/abubakar508/taas/internal/providers/drc/vodacom/airtime"
	"github.com/abubakar508/taas/internal/util"
	"github.com/abubakar508/taas/internal/web"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AirtimeHandler struct {
	service *airtime.Service
}

func NewAirtimeHandler(pool *pgxpool.Pool) *AirtimeHandler {
	return &AirtimeHandler{
		service: airtime.NewService(pool),
	}
}

func (h *AirtimeHandler) Register(r fiber.Router) {
	r.Post("/airtime/purchase", h.purchase)
}

func (h *AirtimeHandler) purchase(c *fiber.Ctx) error {
	partner, err := web.GetPartnerFromCtx(c)
	if err != nil {
		return web.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	idemKey := c.Get("Idempotency-Key")
	if idemKey == "" {
		return web.WriteError(c, fiber.StatusBadRequest, "Idempotency-Key header is required")
	}

	var req airtime.PurchaseRequest
	if err := c.BodyParser(&req); err != nil {
		return web.WriteError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.MSISDN == "" || req.Amount <= 0 || req.Currency == "" {
		return web.WriteError(c, fiber.StatusBadRequest, "msisdn, amount and currency are required")
	}

	if !util.IsValidDRCVodacomMSISDN(req.MSISDN) {
		return web.WriteError(c, fiber.StatusBadRequest, "invalid vodacom drc msisdn format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	resp, err := h.service.Purchase(ctx, partner.ID, idemKey, req)
	if err != nil {
		return web.WriteError(c, fiber.StatusBadGateway, err.Error())
	}

	return c.JSON(resp)
}
