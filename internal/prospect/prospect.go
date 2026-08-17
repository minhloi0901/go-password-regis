package prospect

import "time"

type Prospect struct {
	ID               string
	Username         string
	Email            string
	Status           string
	CreatedAt        time.Time
	ExpiresAt        time.Time
	VerificationCode string
	CodeExpiresAt    time.Time
}

const (
	StatusPending   = "pending"
	StatusActive    = "active"
	StatusSuspended = "suspended"

	ProspectTTL = time.Hour

	VerificationCodeLength = 6
	VerificationCodeDigits = "0123456789"
	VerificationCodeTTL    = 5 * time.Minute
)
