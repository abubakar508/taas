package credentials

import "encoding/json"

type Credential struct {
	ID        string
	PartnerID string
	Country   string
	Provider  string
	Meta      json.RawMessage
	IsActive  bool
}
