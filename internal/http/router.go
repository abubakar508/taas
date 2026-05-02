package http

import (
	"time"

	drcroutes "github.com/abubakar508/taas/internal/countries/drc"
	"github.com/abubakar508/taas/internal/domain/partners"
	"github.com/abubakar508/taas/internal/logger"
	appmw "github.com/abubakar508/taas/internal/middleware"
	"github.com/abubakar508/taas/internal/web"
	"github.com/gofiber/fiber/v2"
	fiberrecover "github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/gofiber/fiber/v2/middleware/timeout"
	"github.com/jackc/pgx/v5/pgxpool"
)

func NewRouter(pool *pgxpool.Pool, log *logger.Logger) *fiber.App {
	app := fiber.New(fiber.Config{
		AppName: "taas-api",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return web.WriteError(c, fiber.StatusInternalServerError, err.Error())
		},
	})

	app.Use(fiberrecover.New())
	app.Use(appmw.RequestID())
	app.Use(appmw.RequestLogger(log))

	app.Get("/healthz", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})

	app.Get("/readyz", func(c *fiber.Ctx) error {
		if err := pool.Ping(c.Context()); err != nil {
			return web.WriteError(c, fiber.StatusServiceUnavailable, "database not ready")
		}
		return c.JSON(fiber.Map{"status": "ready"})
	})

	api := app.Group("/api")

	partnersRepo := partners.NewRepository(pool)
	authMw := web.NewAuthMiddleware(partnersRepo)

	drcGroup := api.Group("/drc")
	drcGroup.Use(authMw.RequirePartnerAuth())
	drcGroup.Use(timeout.NewWithContext(func(c *fiber.Ctx) error {
		return c.Next()
	}, 35*time.Second))

	drcroutes.RegisterRoutes(drcGroup, pool)

	return app
}
