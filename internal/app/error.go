package app

import (
	"errors"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrEmailAlreadyUsed  = errors.New("email already used")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrTokenInvalid      = errors.New("token invalid")
	ErrSameNewPassword   = errors.New("same new password")
)
