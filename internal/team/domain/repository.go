package domain

import (
	"context"
)

type TeamRepository interface {
	Create(ctx context.Context, t *Team) error
	GetByName(ctx context.Context, name string) (*Team, error)
	List(ctx context.Context) ([]*Team, error)
}
