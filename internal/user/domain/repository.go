package domain

import "context"

type UserRepository interface {
	AddTeamMembers(ctx context.Context, teamName string, members []User) error
	GetByID(ctx context.Context, id string) (*User, error)
	ListByTeam(ctx context.Context, teamName string) ([]*User, error)
	UpdateActive(ctx context.Context, id string, active bool) error
	DeactivateByTeam(ctx context.Context, teamName string) error
}
