package prospect

import (
	"errors"
)

var (
	ErrEmailTaken       = errors.New("email already used")
	ErrUsernameTaken    = errors.New("username already used")
	ErrProspectConflict = errors.New("prospect already exists")
	ErrProspectNotFound = errors.New("prospect not found")
	ErrInvalidCode      = errors.New("verification code is invalid")
	ErrCodeExpired      = errors.New("verification code has expired")
)
