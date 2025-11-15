package domain

import "errors"

var (
	ErrInternalDatabase = errors.New("user: internal database error")
	ErrUserNotFound     = errors.New("user not found")
)
