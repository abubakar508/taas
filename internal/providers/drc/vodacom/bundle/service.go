package bundle

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/abubakar508/taas/internal/domain/transactions"
	vodaclient "github.com/abubakar508/taas/internal/providers/drc/vodacom/client"
	"github.com/abubakar508/taas/internal/providers/drc/vodacom/shared"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type offersAPIResponse struct {
	OutputOffers     []offerItem `json:"outputOffers"`
	SessionID        string      `json:"sessionId"`
	TransactionID    string      `json:"transactionId"`
	EventDetail      string      `json:"eventDetail"`
	EventDescription string      `json:"eventDescription"`
}

type offerItem struct {
	Key               string      `json:"key"`
	BundleDesc        string      `json:"bundleDesc"`
	AmountUnconverted json.Number `json:"amountUnconverted"`
	BundleID          any         `json:"bundle_id"`
}

type allocationAPIResponse struct {
	OutputResponseCode string `json:"outputResponseCode"`
	EventDescription   string `json:"eventDescription"`
	TransactionID      string `json:"transactionId"`
}

type Service struct {
	client *vodaclient.Client
	txRepo transactions.Repository
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{
		client: vodaclient.New(pool),
		txRepo: transactions.NewRepository(pool),
	}
}

func (s *Service) ListOffers(ctx context.Context, partnerID, msisdn string) (*ListResponse, error) {
	meta, err := s.client.LoadMeta(ctx, partnerID)
	if err != nil {
		return nil, err
	}

	token, err := s.client.GetAccessToken(ctx, meta)
	if err != nil {
		return nil, err
	}

	url := meta.BaseURL + "/api/v1/airtime/distributor/getOffer/request"

	payload := map[string]any{
		"username":    meta.OfferUsername,
		"password":    meta.OfferPassword,
		"action":      "transact",
		"Flag":        "get_offers",
		"cust_msisdn": msisdn,
		"Language":    "EN",
	}

	raw, statusCode, err := s.client.PostJSON(ctx, meta, url, token, payload)
	if err != nil {
		return nil, err
	}
	if statusCode < 200 || statusCode >= 300 {
		return nil, fmt.Errorf("bundle offers request failed: %s", string(raw))
	}

	var apiResp offersAPIResponse
	if err := json.Unmarshal(raw, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse bundle offers response")
	}

	if len(apiResp.OutputOffers) == 0 {
		msg := apiResp.EventDetail
		if msg == "" {
			msg = apiResp.EventDescription
		}
		if msg == "" {
			msg = "no offers returned"
		}
		return nil, fmt.Errorf("%s", msg)
	}

	offers := make([]Offer, 0, len(apiResp.OutputOffers))
	for _, item := range apiResp.OutputOffers {
		amt, _ := item.AmountUnconverted.Float64()
		offers = append(offers, Offer{
			Key:               item.Key,
			BundleDesc:        item.BundleDesc,
			AmountUnconverted: amt,
			BundleID:          shared.NormalizeBundleID(item.BundleID),
		})
	}

	return &ListResponse{
		Status:                "success",
		Message:               "bundle offers fetched successfully",
		ProviderSessionID:     apiResp.SessionID,
		ProviderTransactionID: apiResp.TransactionID,
		EventDetail:           apiResp.EventDetail,
		EventDescription:      apiResp.EventDescription,
		Offers:                offers,
	}, nil
}

func (s *Service) Purchase(ctx context.Context, partnerID, idemKey string, req PurchaseRequest) (*PurchaseResponse, error) {
	if idemKey == "" {
		return nil, fmt.Errorf("idempotency key required")
	}

	existing, err := s.txRepo.FindByIdempotencyKey(ctx, partnerID, idemKey)
	if err == nil && existing != nil {
		return &PurchaseResponse{
			Status:             existing.Status,
			Message:            "request already processed",
			ProviderRef:        shared.Deref(existing.ProviderRef),
			InternalTxID:       existing.ID,
			ProviderTxID:       shared.Deref(existing.ProviderRef),
			ProviderStatusCode: "",
			EventDescription:   "",
			IdempotencyKey:     idemKey,
		}, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check idempotency")
	}

	meta, err := s.client.LoadMeta(ctx, partnerID)
	if err != nil {
		return nil, err
	}

	token, err := s.client.GetAccessToken(ctx, meta)
	if err != nil {
		return nil, err
	}

	bundleID := req.BundleID
	tx, err := s.txRepo.Create(ctx, &transactions.Transaction{
		PartnerID:       partnerID,
		Country:         "drc",
		Provider:        "vodacom",
		Type:            "bundle",
		MSISDN:          req.MSISDN,
		BundleID:        &bundleID,
		Amount:          req.Amount,
		Currency:        req.Currency,
		Status:          "pending",
		IdempotencyKey:  shared.StrPtr(idemKey),
		ClientReference: shared.StrPtr(req.ClientReference),
	})
	if err != nil {
		if errors.Is(err, transactions.ErrIdempotencyConflict) {
			existing, lookupErr := s.txRepo.FindByIdempotencyKey(ctx, partnerID, idemKey)
			if lookupErr == nil && existing != nil {
				return &PurchaseResponse{
					Status:             existing.Status,
					Message:            "request already processed",
					ProviderRef:        shared.Deref(existing.ProviderRef),
					InternalTxID:       existing.ID,
					ProviderTxID:       shared.Deref(existing.ProviderRef),
					ProviderStatusCode: "",
					EventDescription:   "",
					IdempotencyKey:     idemKey,
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to create transaction")
	}

	url := meta.BaseURL + "/api/v1/airtime/distributor/allocation/request"

	payload := map[string]any{
		"action":         "transact",
		"flag":           "buy",
		"session_id":     req.ProviderSessionID,
		"cust_msisdn":    req.MSISDN,
		"bundle_id":      req.BundleID,
		"transaction_id": req.ProviderTransactionID,
		"language":       "EN",
	}

	raw, statusCode, err := s.client.PostJSON(ctx, meta, url, token, payload)
	if err != nil {
		msg := "bundle allocation request failed"
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", nil, nil, &msg)
		return nil, fmt.Errorf("%s", msg)
	}
	if statusCode < 200 || statusCode >= 300 {
		msg := fmt.Sprintf("bundle allocation failed: %s", strings.TrimSpace(string(raw)))
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", nil, map[string]any{"raw": string(raw)}, &msg)
		return nil, fmt.Errorf("%s", msg)
	}

	var apiResp allocationAPIResponse
	if err := json.Unmarshal(raw, &apiResp); err != nil {
		msg := "failed to parse allocation response"
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", nil, map[string]any{"raw": string(raw)}, &msg)
		return nil, fmt.Errorf("%s", msg)
	}

	if strings.TrimSpace(apiResp.OutputResponseCode) != "0" {
		msg := apiResp.EventDescription
		if msg == "" {
			msg = "bundle purchase failed"
		}
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", shared.StrPtr(apiResp.TransactionID), apiResp, &msg)
		return nil, fmt.Errorf("%s", msg)
	}

	if err := s.txRepo.UpdateStatus(ctx, tx.ID, "success", shared.StrPtr(apiResp.TransactionID), apiResp, nil); err != nil {
		return nil, fmt.Errorf("failed to update transaction")
	}

	return &PurchaseResponse{
		Status:             "success",
		Message:            "bundle purchased successfully",
		ProviderRef:        apiResp.TransactionID,
		InternalTxID:       tx.ID,
		ProviderTxID:       apiResp.TransactionID,
		ProviderStatusCode: apiResp.OutputResponseCode,
		EventDescription:   apiResp.EventDescription,
		IdempotencyKey:     idemKey,
	}, nil
}
