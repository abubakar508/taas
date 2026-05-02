package transactions

import (
	"encoding/json"
	"time"
)

type Transaction struct {
	ID              string
	PartnerID       string
	Country         string
	Provider        string
	Type            string
	MSISDN          string
	BundleID        *string
	Amount          float64
	Currency        string
	Status          string
	ProviderRef     *string
	IdempotencyKey  *string
	ClientReference *string
	ErrorMessage    *string
	ProviderRaw     json.RawMessage
	CreatedAt       time.Time
	UpdatedAt       time.Time
}
