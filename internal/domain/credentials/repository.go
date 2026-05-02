package credentials

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository interface {
	FindActiveByPartnerCountryProvider(ctx context.Context, partnerID, country, provider string) (*Credential, error)
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) FindActiveByPartnerCountryProvider(ctx context.Context, partnerID, country, provider string) (*Credential, error) {
	const q = `
		SELECT id, partner_id, country, provider, meta, is_active
		FROM partner_provider_credentials
		WHERE partner_id = $1 AND country = $2 AND provider = $3 AND is_active = true
		LIMIT 1
	`

	var c Credential
	err := r.pool.QueryRow(ctx, q, partnerID, country, provider).Scan(
		&c.ID,
		&c.PartnerID,
		&c.Country,
		&c.Provider,
		&c.Meta,
		&c.IsActive,
	)
	if err != nil {
		return nil, errors.New("credential not found")
	}

	return &c, nil
}
