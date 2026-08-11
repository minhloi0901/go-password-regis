package credential

import (
	"errors"
)

var (
	ErrUsernameTaken = errors.New("username already used")
)
