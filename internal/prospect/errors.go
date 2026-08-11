package prospect

import (
	"errors"
)

var (
	ErrEmailTaken    = errors.New("email already used")
	ErrUsernameTaken = errors.New("username already used")
)
