package credential

import (
	"errors"
)

var (
	ErrUsernameTaken      = errors.New("username already used")
	ErrCredentialNotFound = errors.New("credential not found")
)
