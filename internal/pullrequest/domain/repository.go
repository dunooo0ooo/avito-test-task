package domain

import (
	"context"
	"time"
)

type PullRequestRepository interface {
	Create(ctx context.Context, pr *PullRequest) error
	GetByID(ctx context.Context, id string) (*PullRequest, error)
	UpdateStatus(ctx context.Context, id string, status PRStatus, mergedAt *time.Time) error
	SetReviewers(ctx context.Context, id string, reviewerIDs []string) error
	ListByReviewer(ctx context.Context, reviewerID string) ([]PullRequestShort, error)
	ListOpenByReviewers(ctx context.Context, reviewerIDs []string) ([]PullRequestShort, error)
	CountByReviewer(ctx context.Context) (map[string]string, error)
}
