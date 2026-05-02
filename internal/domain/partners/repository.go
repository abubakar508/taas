package partners

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindActiveByAppNameAndKeyHash(ctx context.Context, appName, keyHash string) (*Partner, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) FindActiveByAppNameAndKeyHash(ctx context.Context, appName, keyHash string) (*Partner, error) {
	const q = `
		SELECT id, app_name, key_hash, is_active, created_at, updated_at
		FROM partners
		WHERE app_name = $1 AND key_hash = $2 AND is_active = true
		LIMIT 1
	`

	var p Partner
	err := r.pool.QueryRow(ctx, q, appName, keyHash).Scan(
		&p.ID,
		&p.AppName,
		&p.KeyHash,
		&p.IsActive,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if err != nil {
		return nil, errors.New("partner not found")
	}

	return &p, nil
}
