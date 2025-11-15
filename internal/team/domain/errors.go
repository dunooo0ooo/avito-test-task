package domain

import "errors"

var (
	ErrTeamNotFound      = errors.New("team not found")
	ErrTeamAlreadyExists = errors.New("team already exists")
	ErrInternalDatabase  = errors.New("user: internal database error")
)
