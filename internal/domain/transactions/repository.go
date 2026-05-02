package transactions

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrIdempotencyConflict = errors.New("idempotency conflict")

type Repository interface {
	Create(ctx context.Context, tx *Transaction) (*Transaction, error)
	FindByIdempotencyKey(ctx context.Context, partnerID, idemKey string) (*Transaction, error)
	FindByIDAndPartner(ctx context.Context, id, partnerID string) (*Transaction, error)
	UpdateStatus(ctx context.Context, id, status string, providerRef *string, providerRaw interface{}, errMsg *string) error
}

type repository struct {
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) Repository {
	return &repository{pool: pool}
}

func (r *repository) Create(ctx context.Context, tx *Transaction) (*Transaction, error) {
	id := uuid.NewString()
	tx.ID = id

	const q = `
		INSERT INTO transactions (
			id, partner_id, country, provider, type, msisdn, bundle_id, amount, currency, status,
			provider_ref, idempotency_key, client_reference, error_message, provider_raw
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
		RETURNING created_at, updated_at
	`

	var raw []byte
	if tx.ProviderRaw != nil {
		raw = tx.ProviderRaw
	}

	err := r.pool.QueryRow(
		ctx,
		q,
		tx.ID,
		tx.PartnerID,
		tx.Country,
		tx.Provider,
		tx.Type,
		tx.MSISDN,
		tx.BundleID,
		tx.Amount,
		tx.Currency,
		tx.Status,
		tx.ProviderRef,
		tx.IdempotencyKey,
		tx.ClientReference,
		tx.ErrorMessage,
		raw,
	).Scan(&tx.CreatedAt, &tx.UpdatedAt)

	if err != nil {
		return nil, err
	}

	return tx, nil
}

func (r *repository) FindByIdempotencyKey(ctx context.Context, partnerID, idemKey string) (*Transaction, error) {
	const q = `
		SELECT id, partner_id, country, provider, type, msisdn, bundle_id, amount, currency, status,
		       provider_ref, idempotency_key, client_reference, error_message, provider_raw, created_at, updated_at
		FROM transactions
		WHERE partner_id = $1 AND idempotency_key = $2
		LIMIT 1
	`

	var tx Transaction
	err := r.pool.QueryRow(ctx, q, partnerID, idemKey).Scan(
		&tx.ID,
		&tx.PartnerID,
		&tx.Country,
		&tx.Provider,
		&tx.Type,
		&tx.MSISDN,
		&tx.BundleID,
		&tx.Amount,
		&tx.Currency,
		&tx.Status,
		&tx.ProviderRef,
		&tx.IdempotencyKey,
		&tx.ClientReference,
		&tx.ErrorMessage,
		&tx.ProviderRaw,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (r *repository) FindByIDAndPartner(ctx context.Context, id, partnerID string) (*Transaction, error) {
	const q = `
		SELECT id, partner_id, country, provider, type, msisdn, bundle_id, amount, currency, status,
		       provider_ref, idempotency_key, client_reference, error_message, provider_raw, created_at, updated_at
		FROM transactions
		WHERE id = $1 AND partner_id = $2
		LIMIT 1
	`

	var tx Transaction
	err := r.pool.QueryRow(ctx, q, id, partnerID).Scan(
		&tx.ID,
		&tx.PartnerID,
		&tx.Country,
		&tx.Provider,
		&tx.Type,
		&tx.MSISDN,
		&tx.BundleID,
		&tx.Amount,
		&tx.Currency,
		&tx.Status,
		&tx.ProviderRef,
		&tx.IdempotencyKey,
		&tx.ClientReference,
		&tx.ErrorMessage,
		&tx.ProviderRaw,
		&tx.CreatedAt,
		&tx.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (r *repository) UpdateStatus(ctx context.Context, id, status string, providerRef *string, providerRaw interface{}, errMsg *string) error {
	var raw []byte
	if providerRaw != nil {
		b, _ := json.Marshal(providerRaw)
		raw = b
	}

	const q = `
		UPDATE transactions
		SET status = $2,
		    provider_ref = COALESCE($3, provider_ref),
		    provider_raw = COALESCE($4, provider_raw),
		    error_message = $5,
		    updated_at = NOW()
		WHERE id = $1
	`

	_, err := r.pool.Exec(ctx, q, id, status, providerRef, raw, errMsg)
	return err
}
