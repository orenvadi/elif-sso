package storage

import "errors"

var (
	ErrUserExists          = errors.New("user already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrAppNotFound         = errors.New("app not found")
	ErrConfirmCodeNotFound = errors.New("confirm code not found")
)
