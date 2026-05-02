package vodacom

import (
	"context"
	"time"

	"github.com/abubakar508/taas/internal/providers/drc/vodacom/bundle"
	"github.com/abubakar508/taas/internal/util"
	"github.com/abubakar508/taas/internal/web"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BundleHandler struct {
	service *bundle.Service
}

func NewBundleHandler(pool *pgxpool.Pool) *BundleHandler {
	return &BundleHandler{
		service: bundle.NewService(pool),
	}
}

func (h *BundleHandler) Register(r fiber.Router) {
	r.Post("/bundles/offers", h.listOffers)
	r.Post("/bundles/purchase", h.purchase)
}

func (h *BundleHandler) listOffers(c *fiber.Ctx) error {
	partner, err := web.GetPartnerFromCtx(c)
	if err != nil {
		return web.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	var req bundle.ListRequest
	if err := c.BodyParser(&req); err != nil {
		return web.WriteError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.MSISDN == "" {
		return web.WriteError(c, fiber.StatusBadRequest, "msisdn is required")
	}

	if !util.IsValidDRCVodacomMSISDN(req.MSISDN) {
		return web.WriteError(c, fiber.StatusBadRequest, "invalid vodacom drc msisdn format")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	resp, err := h.service.ListOffers(ctx, partner.ID, req.MSISDN)
	if err != nil {
		return web.WriteError(c, fiber.StatusBadGateway, "failed to fetch bundle offers")
	}

	return c.JSON(resp)
}

func (h *BundleHandler) purchase(c *fiber.Ctx) error {
	partner, err := web.GetPartnerFromCtx(c)
	if err != nil {
		return web.WriteError(c, fiber.StatusUnauthorized, "unauthorized")
	}

	idemKey := c.Get("Idempotency-Key")
	if idemKey == "" {
		return web.WriteError(c, fiber.StatusBadRequest, "Idempotency-Key header is required")
	}

	var req bundle.PurchaseRequest
	if err := c.BodyParser(&req); err != nil {
		return web.WriteError(c, fiber.StatusBadRequest, "invalid request body")
	}

	if req.MSISDN == "" || req.BundleID == "" || req.Amount <= 0 || req.Currency == "" || req.ProviderSessionID == "" || req.ProviderTransactionID == "" {
		return web.WriteError(c, fiber.StatusBadRequest, "msisdn, bundle_id, amount, currency, provider_session_id and provider_transaction_id are required")
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
