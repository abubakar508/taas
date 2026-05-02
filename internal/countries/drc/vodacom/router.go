package vodacom

import (
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

func RegisterRoutes(r fiber.Router, pool *pgxpool.Pool) {
	NewBundleHandler(pool).Register(r)
	NewAirtimeHandler(pool).Register(r)
	NewTransactionHandler(pool).Register(r)
}
