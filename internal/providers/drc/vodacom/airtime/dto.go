package airtime

type PurchaseRequest struct {
	MSISDN          string  `json:"msisdn"`
	Amount          float64 `json:"amount"`
	Currency        string  `json:"currency"`
	ClientReference string  `json:"client_reference"`
}

type PurchaseResponse struct {
	Status         string `json:"status"`
	Message        string `json:"message"`
	ProviderRef    string `json:"provider_ref,omitempty"`
	TransactionID  string `json:"transaction_id"`
	IdempotencyKey string `json:"idempotency_key,omitempty"`
}
