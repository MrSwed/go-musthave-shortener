package errors

import (
	"errors"
)

var (
	ErrNotExist     = errors.New("does not exist")
	ErrAlreadyExist = errors.New("already exist")
)
