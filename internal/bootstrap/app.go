package bootstrap

import (
	"context"

	"github.com/abubakar508/taas/internal/config"
	"github.com/abubakar508/taas/internal/db"
	"github.com/abubakar508/taas/internal/http"
	"github.com/abubakar508/taas/internal/logger"
	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type App struct {
	Config *config.Config
	Logger *logger.Logger
	DB     *pgxpool.Pool
	HTTP   *fiber.App
}

func New(ctx context.Context) (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	log := logger.New()

	pool, err := db.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	app := http.NewRouter(pool, log)

	return &App{
		Config: cfg,
		Logger: log,
		DB:     pool,
		HTTP:   app,
	}, nil
}
