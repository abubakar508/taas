package airtime

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/abubakar508/taas/internal/domain/transactions"
	vodaclient "github.com/abubakar508/taas/internal/providers/drc/vodacom/client"
	"github.com/abubakar508/taas/internal/providers/drc/vodacom/shared"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type soapFaultEnvelope struct {
	Body soapFaultBody `xml:"Body"`
}

type soapFaultBody struct {
	Fault *soapFault `xml:"Fault"`
}

type soapFault struct {
	FaultCode   string `xml:"faultcode"`
	FaultString string `xml:"faultstring"`
}

type soapEnvelope struct {
	XMLName xml.Name  `xml:"soapenv:Envelope"`
	Soapenv string    `xml:"xmlns:soapenv,attr"`
	API     string    `xml:"xmlns:api,attr"`
	Req     string    `xml:"xmlns:req,attr"`
	Com     string    `xml:"xmlns:com,attr"`
	Header  *struct{} `xml:"soapenv:Header"`
	Body    soapBody  `xml:"soapenv:Body"`
}

type soapBody struct {
	Request soapAPIRequest `xml:"api:Request"`
}

type soapAPIRequest struct {
	Header soapRequestHeader `xml:"req:Header"`
	Body   soapRequestBody   `xml:"req:Body"`
}

type soapRequestHeader struct {
	Version                  string     `xml:"req:Version"`
	CommandID                string     `xml:"req:CommandID"`
	OriginatorConversationID string     `xml:"req:OriginatorConversationID"`
	Caller                   soapCaller `xml:"req:Caller"`
	KeyOwner                 string     `xml:"req:KeyOwner"`
	Timestamp                string     `xml:"req:Timestamp"`
}

type soapCaller struct {
	CallerType   string `xml:"req:CallerType"`
	ThirdPartyID string `xml:"req:ThirdPartyID"`
	Password     string `xml:"req:Password"`
}

type soapRequestBody struct {
	Identity           soapIdentity           `xml:"req:Identity"`
	TransactionRequest soapTransactionRequest `xml:"req:TransactionRequest"`
}

type soapIdentity struct {
	Initiator soapInitiator `xml:"req:Initiator"`
}

type soapInitiator struct {
	IdentifierType     string `xml:"req:IdentifierType"`
	Identifier         string `xml:"req:Identifier"`
	SecurityCredential string `xml:"req:SecurityCredential"`
	ShortCode          string `xml:"req:ShortCode"`
}

type soapTransactionRequest struct {
	Parameters soapParameters `xml:"req:Parameters"`
}

type soapParameters struct {
	Parameter soapParameter `xml:"req:Parameter"`
	Amount    string        `xml:"req:Amount"`
	Currency  string        `xml:"req:Currency"`
}

type soapParameter struct {
	Key   string `xml:"com:Key"`
	Value string `xml:"com:Value"`
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

func (s *Service) Purchase(ctx context.Context, partnerID, idemKey string, req PurchaseRequest) (*PurchaseResponse, error) {
	if idemKey == "" {
		return nil, fmt.Errorf("idempotency key required")
	}

	existing, err := s.txRepo.FindByIdempotencyKey(ctx, partnerID, idemKey)
	if err == nil && existing != nil {
		return &PurchaseResponse{
			Status:         existing.Status,
			Message:        "request already processed",
			ProviderRef:    shared.Deref(existing.ProviderRef),
			TransactionID:  existing.ID,
			IdempotencyKey: idemKey,
		}, nil
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("failed to check idempotency")
	}

	meta, err := s.client.LoadMeta(ctx, partnerID)
	if err != nil {
		return nil, err
	}

	tx, err := s.txRepo.Create(ctx, &transactions.Transaction{
		PartnerID:       partnerID,
		Country:         "drc",
		Provider:        "vodacom",
		Type:            "airtime",
		MSISDN:          req.MSISDN,
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
					Status:         existing.Status,
					Message:        "request already processed",
					ProviderRef:    shared.Deref(existing.ProviderRef),
					TransactionID:  existing.ID,
					IdempotencyKey: idemKey,
				}, nil
			}
		}
		return nil, fmt.Errorf("failed to create transaction")
	}

	now := time.Now()
	envelope := soapEnvelope{
		Soapenv: "http://schemas.xmlsoap.org/soap/envelope/",
		API:     "http://cps.huawei.com/synccpsinterface/api_requestmgr",
		Req:     "http://cps.huawei.com/synccpsinterface/request",
		Com:     "http://cps.huawei.com/synccpsinterface/common",
		Header:  &struct{}{},
		Body: soapBody{
			Request: soapAPIRequest{
				Header: soapRequestHeader{
					Version:                  "1.0",
					CommandID:                meta.CommandID,
					OriginatorConversationID: "RZO-" + now.Format("20060102150405"),
					Caller: soapCaller{
						CallerType:   meta.CallerType,
						ThirdPartyID: meta.ThirdPartyID,
						Password:     meta.ThirdPartyPassword,
					},
					KeyOwner:  meta.KeyOwner,
					Timestamp: now.Format("20060102150405"),
				},
				Body: soapRequestBody{
					Identity: soapIdentity{
						Initiator: soapInitiator{
							IdentifierType:     "11",
							Identifier:         meta.InitiatorIdentifier,
							SecurityCredential: meta.SecurityCredential,
							ShortCode:          meta.ShortCode,
						},
					},
					TransactionRequest: soapTransactionRequest{
						Parameters: soapParameters{
							Parameter: soapParameter{
								Key:   "RechargedMSISDN",
								Value: req.MSISDN,
							},
							Amount:   fmt.Sprintf("%.2f", req.Amount),
							Currency: req.Currency,
						},
					},
				},
			},
		},
	}

	payload, err := vodaclient.MarshalSOAP(envelope)
	if err != nil {
		msg := "failed to build soap request"
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", nil, nil, &msg)
		return nil, fmt.Errorf("%s", msg)
	}

	raw, statusCode, err := s.client.PostSOAP(ctx, meta, payload)
	if err != nil {
		msg := "airtime request failed"
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", nil, nil, &msg)
		return nil, fmt.Errorf("%s", msg)
	}

	rawStr := string(raw)

	if statusCode < 200 || statusCode >= 300 {
		msg := "airtime request failed"
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", nil, map[string]any{"raw": rawStr}, &msg)
		return nil, fmt.Errorf("%s", msg)
	}

	var faultEnvelope soapFaultEnvelope
	_ = xml.Unmarshal(raw, &faultEnvelope)
	if faultEnvelope.Body.Fault != nil {
		msg := faultEnvelope.Body.Fault.FaultString
		if msg == "" {
			msg = "soap fault"
		}
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", nil, map[string]any{"raw": rawStr}, &msg)
		return nil, fmt.Errorf("%s", msg)
	}

	if strings.TrimSpace(rawStr) == "" {
		msg := "empty airtime response"
		_ = s.txRepo.UpdateStatus(ctx, tx.ID, "failed", nil, nil, &msg)
		return nil, fmt.Errorf("%s", msg)
	}

	if err := s.txRepo.UpdateStatus(ctx, tx.ID, "success", nil, map[string]any{"raw": rawStr}, nil); err != nil {
		return nil, fmt.Errorf("failed to update transaction")
	}

	return &PurchaseResponse{
		Status:         "success",
		Message:        "airtime purchase accepted",
		TransactionID:  tx.ID,
		IdempotencyKey: idemKey,
	}, nil
}
