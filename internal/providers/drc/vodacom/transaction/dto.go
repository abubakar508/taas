package transaction

import "encoding/json"

type LookupResponse struct {
	ID              string          `json:"id"`
	Country         string          `json:"country"`
	Provider        string          `json:"provider"`
	Type            string          `json:"type"`
	MSISDN          string          `json:"msisdn"`
	BundleID        *string         `json:"bundle_id,omitempty"`
	Amount          float64         `json:"amount"`
	Currency        string          `json:"currency"`
	Status          string          `json:"status"`
	ProviderRef     *string         `json:"provider_ref,omitempty"`
	ErrorMessage    *string         `json:"error_message,omitempty"`
	IdempotencyKey  *string         `json:"idempotency_key,omitempty"`
	ClientReference *string         `json:"client_reference,omitempty"`
	ProviderRaw     json.RawMessage `json:"provider_raw,omitempty"`
	CreatedAt       string          `json:"created_at"`
	UpdatedAt       string          `json:"updated_at"`
}
