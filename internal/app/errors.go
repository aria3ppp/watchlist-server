package app

import (
	"errors"
)

var (
	ErrNotFound          = errors.New("not found")
	ErrUsedEmail         = errors.New("email used")
	ErrIncorrectPassword = errors.New("incorrect password")
	ErrSamePassword      = errors.New("same password")
)
