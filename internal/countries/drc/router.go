package drc

import (
	drcvodacom "github.com/abubakar508/taas/internal/countries/drc/vodacom"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(r fiber.Router, pool *pgxpool.Pool) {
	drcvodacom.RegisterRoutes(r.Group("/vodacom"), pool)
}
