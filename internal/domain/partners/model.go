package partners

import "time"

type Partner struct {
	ID        string
	AppName   string
	KeyHash   string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}
