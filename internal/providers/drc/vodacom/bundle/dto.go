package bundle

type Offer struct {
	Key               string  `json:"key"`
	BundleDesc        string  `json:"bundle_desc"`
	AmountUnconverted float64 `json:"amount_unconverted"`
	BundleID          string  `json:"bundle_id"`
}

type ListRequest struct {
	MSISDN string `json:"msisdn"`
}

type ListResponse struct {
	Status                string  `json:"status"`
	Message               string  `json:"message"`
	ProviderSessionID     string  `json:"provider_session_id,omitempty"`
	ProviderTransactionID string  `json:"provider_transaction_id,omitempty"`
	EventDetail           string  `json:"event_detail,omitempty"`
	EventDescription      string  `json:"event_description,omitempty"`
	Offers                []Offer `json:"offers"`
}

type PurchaseRequest struct {
	MSISDN                string  `json:"msisdn"`
	BundleID              string  `json:"bundle_id"`
	Amount                float64 `json:"amount"`
	Currency              string  `json:"currency"`
	ProviderSessionID     string  `json:"provider_session_id"`
	ProviderTransactionID string  `json:"provider_transaction_id"`
	ClientReference       string  `json:"client_reference"`
}

type PurchaseResponse struct {
	Status             string `json:"status"`
	Message            string `json:"message"`
	ProviderRef        string `json:"provider_ref,omitempty"`
	InternalTxID       string `json:"internal_transaction_id"`
	ProviderTxID       string `json:"provider_transaction_id,omitempty"`
	ProviderStatusCode string `json:"provider_status_code,omitempty"`
	EventDescription   string `json:"event_description,omitempty"`
	IdempotencyKey     string `json:"idempotency_key,omitempty"`
}
