package transaction

import (
	"context"
	"fmt"
	"time"

	"github.com/abubakar508/taas/internal/domain/transactions"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Service struct {
	txRepo transactions.Repository
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		txRepo: transactions.NewRepository(pool),
	}
}

func (s *Service) Get(ctx context.Context, partnerID, txID string) (*LookupResponse, error) {
	tx, err := s.txRepo.FindByIDAndPartner(ctx, txID, partnerID)
	if err != nil {
		return nil, fmt.Errorf("transaction not found")
	}

	return &LookupResponse{
		ID:              tx.ID,
		Country:         tx.Country,
		Provider:        tx.Provider,
		Type:            tx.Type,
		MSISDN:          tx.MSISDN,
		BundleID:        tx.BundleID,
		Amount:          tx.Amount,
		Currency:        tx.Currency,
		Status:          tx.Status,
		ProviderRef:     tx.ProviderRef,
		ErrorMessage:    tx.ErrorMessage,
		IdempotencyKey:  tx.IdempotencyKey,
		ClientReference: tx.ClientReference,
		ProviderRaw:     tx.ProviderRaw,
		CreatedAt:       tx.CreatedAt.Format(time.RFC3339),
		UpdatedAt:       tx.UpdatedAt.Format(time.RFC3339),
	}, nil
}
